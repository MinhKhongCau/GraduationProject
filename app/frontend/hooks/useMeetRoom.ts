"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getChatSocket } from "@/lib/chatSocket";

export type MeetRole = "PATIENT" | "EXPERT";

export interface MeetPresence {
  presenceSeconds: number;
  requiredSeconds: number;
  remainingSeconds: number;
  canComplete: boolean;
}

export type MeetStatus =
  | "preparing" // asking for camera/mic
  | "joining" // waiting for the socket / meet:join ack
  | "waiting" // in the room, the other participant is not here yet
  | "connecting" // peer is here, WebRTC negotiating
  | "connected"
  | "ended"
  | "error";

export interface MeetError {
  code: string;
  message: string;
}

interface JoinAck {
  ok: boolean;
  code?: string;
  message?: string;
  appointmentId?: string;
  role?: MeetRole;
  startTime?: number;
  endTime?: number;
  presence?: MeetPresence;
}

interface EndAck {
  ok: boolean;
  code?: string;
  message?: string;
  retryable?: boolean;
  settled?: boolean;
  presence?: MeetPresence;
}

interface SignalMessage {
  from: string;
  role: MeetRole;
  type: "offer" | "answer" | "ice";
  payload: unknown;
}

const ICE_SERVERS: RTCIceServer[] = [
  { urls: "stun:stun.l.google.com:19302" },
  { urls: "stun:stun1.l.google.com:19302" },
];

/**
 * Peer-to-peer call for a paid appointment, joined through the meeting link
 * booking-service generates (…/meet/<meetingToken>). chatroom-service's
 * meet.socket.controller.js only authorizes the participants and relays
 * `meet:signal`; media flows directly between the two browsers.
 *
 * Negotiation: whoever is already in the room makes the offer when the other
 * participant joins (meet:participant-joined), so a late/rejoining peer always
 * gets a fresh connection.
 */
export function useMeetRoom(meetingToken: string) {
  const socket = useMemo(() => getChatSocket(), []);

  const [status, setStatus] = useState<MeetStatus>("preparing");
  const [error, setError] = useState<MeetError | null>(null);
  const [role, setRole] = useState<MeetRole | null>(null);
  const [startTime, setStartTime] = useState<number | null>(null);
  const [endTime, setEndTime] = useState<number | null>(null);
  const [presence, setPresence] = useState<MeetPresence | null>(null);
  const [canComplete, setCanComplete] = useState(false);
  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStream, setRemoteStream] = useState<MediaStream | null>(null);
  const [isMuted, setIsMuted] = useState(false);
  const [isVideoOff, setIsVideoOff] = useState(false);
  const [isEnding, setIsEnding] = useState(false);

  const localStreamRef = useRef<MediaStream | null>(null);
  const pcRef = useRef<RTCPeerConnection | null>(null);
  const pendingIceRef = useRef<RTCIceCandidateInit[]>([]);
  const joinedRef = useRef(false);
  const endedRef = useRef(false);

  const closePeer = useCallback(() => {
    const pc = pcRef.current;
    pcRef.current = null;
    pendingIceRef.current = [];
    if (pc) {
      pc.onicecandidate = null;
      pc.ontrack = null;
      pc.onconnectionstatechange = null;
      pc.close();
    }
    setRemoteStream(null);
  }, []);

  const createPeer = useCallback(() => {
    closePeer();
    const pc = new RTCPeerConnection({ iceServers: ICE_SERVERS });
    pcRef.current = pc;

    localStreamRef.current?.getTracks().forEach((track) => pc.addTrack(track, localStreamRef.current!));

    pc.onicecandidate = (event) => {
      if (event.candidate) socket.emit("meet:signal", { type: "ice", payload: event.candidate.toJSON() });
    };
    pc.ontrack = (event) => {
      const [stream] = event.streams;
      if (stream) setRemoteStream(stream);
    };
    pc.onconnectionstatechange = () => {
      if (pcRef.current !== pc || endedRef.current) return;
      if (pc.connectionState === "connected") setStatus("connected");
      else if (pc.connectionState === "failed" || pc.connectionState === "disconnected") setStatus("connecting");
    };
    return pc;
  }, [socket, closePeer]);

  const flushPendingIce = useCallback(async (pc: RTCPeerConnection) => {
    const queued = pendingIceRef.current;
    pendingIceRef.current = [];
    for (const candidate of queued) {
      try {
        await pc.addIceCandidate(candidate);
      } catch (err) {
        console.warn("meet: dropped ICE candidate", err);
      }
    }
  }, []);

  const makeOffer = useCallback(async () => {
    const pc = createPeer();
    setStatus("connecting");
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    socket.emit("meet:signal", { type: "offer", payload: offer });
  }, [createPeer, socket]);

  // 1) Camera + microphone (falls back to audio-only when there is no camera).
  useEffect(() => {
    let cancelled = false;
    (async () => {
      let stream: MediaStream | null = null;
      try {
        stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
      } catch {
        try {
          stream = await navigator.mediaDevices.getUserMedia({ audio: true });
          setIsVideoOff(true);
        } catch {
          stream = null;
        }
      }
      if (cancelled) {
        stream?.getTracks().forEach((t) => t.stop());
        return;
      }
      if (!stream) {
        setError({ code: "no_media", message: "Allow camera or microphone access to join the meeting." });
        setStatus("error");
        return;
      }
      localStreamRef.current = stream;
      setLocalStream(stream);
      setStatus("joining");
    })();
    return () => {
      cancelled = true;
      localStreamRef.current?.getTracks().forEach((t) => t.stop());
      localStreamRef.current = null;
    };
  }, []);

  // 2) Join the meet room once media is ready, and rejoin after every reconnect
  //    (the server forgets socket.data.meet when the socket changes).
  const mediaReady = localStream !== null;
  useEffect(() => {
    if (!mediaReady || !meetingToken) return;

    const join = () => {
      if (endedRef.current) return;
      setStatus((s) => (s === "connected" ? s : "joining"));
      socket.emit("meet:join", { meetingToken }, (ack: JoinAck) => {
        if (!ack?.ok) {
          joinedRef.current = false;
          setError({ code: ack?.code ?? "unavailable", message: ack?.message ?? "Could not join the meeting." });
          setStatus(ack?.code === "ended" ? "ended" : "error");
          if (ack?.code === "ended") endedRef.current = true;
          return;
        }
        joinedRef.current = true;
        setError(null);
        setRole(ack.role ?? null);
        setStartTime(ack.startTime ?? null);
        setEndTime(ack.endTime ?? null);
        if (ack.presence) {
          setPresence(ack.presence);
          setCanComplete(ack.presence.canComplete);
        }
        // Wait here: the participant already in the room sends the offer.
        if (!pcRef.current) setStatus("waiting");
      });
    };

    const onConnect = () => join();
    const onDisconnect = () => {
      joinedRef.current = false;
      closePeer();
      if (!endedRef.current) setStatus("joining");
    };

    const onParticipantJoined = () => {
      makeOffer().catch((err) => console.error("meet: offer failed", err));
    };
    const onParticipantLeft = () => {
      closePeer();
      if (!endedRef.current) setStatus("waiting");
    };

    const onSignal = async (msg: SignalMessage) => {
      try {
        if (msg.type === "offer") {
          const pc = createPeer();
          setStatus("connecting");
          await pc.setRemoteDescription(msg.payload as RTCSessionDescriptionInit);
          await flushPendingIce(pc);
          const answer = await pc.createAnswer();
          await pc.setLocalDescription(answer);
          socket.emit("meet:signal", { type: "answer", payload: answer });
        } else if (msg.type === "answer") {
          const pc = pcRef.current;
          if (!pc || pc.signalingState !== "have-local-offer") return;
          await pc.setRemoteDescription(msg.payload as RTCSessionDescriptionInit);
          await flushPendingIce(pc);
        } else if (msg.type === "ice") {
          const pc = pcRef.current;
          const candidate = msg.payload as RTCIceCandidateInit;
          if (pc?.remoteDescription) await pc.addIceCandidate(candidate);
          else pendingIceRef.current.push(candidate);
        }
      } catch (err) {
        console.error("meet: signal handling failed", err);
      }
    };

    const onCanComplete = (data: { presence?: MeetPresence }) => {
      setCanComplete(true);
      if (data?.presence) setPresence(data.presence);
    };
    const onEnded = () => {
      endedRef.current = true;
      joinedRef.current = false;
      closePeer();
      setStatus("ended");
    };

    socket.on("connect", onConnect);
    socket.on("disconnect", onDisconnect);
    socket.on("meet:participant-joined", onParticipantJoined);
    socket.on("meet:participant-left", onParticipantLeft);
    socket.on("meet:signal", onSignal);
    socket.on("meet:can-complete", onCanComplete);
    socket.on("meet:ended", onEnded);

    // ChatProvider owns the connection; just make sure it is up.
    if (socket.connected) join();
    else socket.connect();

    return () => {
      socket.off("connect", onConnect);
      socket.off("disconnect", onDisconnect);
      socket.off("meet:participant-joined", onParticipantJoined);
      socket.off("meet:participant-left", onParticipantLeft);
      socket.off("meet:signal", onSignal);
      socket.off("meet:can-complete", onCanComplete);
      socket.off("meet:ended", onEnded);
      if (joinedRef.current) socket.emit("meet:leave");
      joinedRef.current = false;
      closePeer();
    };
  }, [mediaReady, meetingToken, socket, closePeer, createPeer, flushPendingIce, makeOffer]);

  // Expert only: refresh the presence counter while waiting for the minimum.
  useEffect(() => {
    if (role !== "EXPERT" || canComplete || status === "ended" || status === "error") return;
    const timer = setInterval(() => {
      if (!joinedRef.current) return;
      socket.emit("meet:presence", null, (ack: { ok: boolean; presence?: MeetPresence }) => {
        if (ack?.ok && ack.presence) {
          setPresence(ack.presence);
          if (ack.presence.canComplete) setCanComplete(true);
        }
      });
    }, 15_000);
    return () => clearInterval(timer);
  }, [role, canComplete, status, socket]);

  const toggleMute = useCallback(() => {
    const tracks = localStreamRef.current?.getAudioTracks() ?? [];
    const next = !isMuted;
    tracks.forEach((t) => (t.enabled = !next));
    setIsMuted(next);
  }, [isMuted]);

  const toggleVideo = useCallback(() => {
    const tracks = localStreamRef.current?.getVideoTracks() ?? [];
    if (tracks.length === 0) return;
    const next = !isVideoOff;
    tracks.forEach((t) => (t.enabled = !next));
    setIsVideoOff(next);
  }, [isVideoOff]);

  const leave = useCallback(() => {
    if (joinedRef.current) socket.emit("meet:leave");
    joinedRef.current = false;
    closePeer();
    localStreamRef.current?.getTracks().forEach((t) => t.stop());
  }, [socket, closePeer]);

  /** Expert only: completes the session (and triggers the payout) once presence ≥ minimum. */
  const endSession = useCallback(
    () =>
      new Promise<EndAck>((resolve) => {
        setIsEnding(true);
        socket.emit("meet:end", null, (ack: EndAck) => {
          setIsEnding(false);
          if (ack?.presence) setPresence(ack.presence);
          resolve(ack ?? { ok: false, code: "unavailable", message: "No response from the server" });
        });
      }),
    [socket]
  );

  return {
    status,
    error,
    role,
    startTime,
    endTime,
    presence,
    canComplete,
    localStream,
    remoteStream,
    isMuted,
    isVideoOff,
    hasCamera: (localStream?.getVideoTracks().length ?? 0) > 0,
    isEnding,
    toggleMute,
    toggleVideo,
    leave,
    endSession,
  };
}
