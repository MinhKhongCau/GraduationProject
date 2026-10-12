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
  /** Sender's socket id; signals are addressed back to it with `to`. */
  fromSocketId?: string;
  role: MeetRole;
  type: "offer" | "answer" | "ice";
  payload: unknown;
}

interface ParticipantEvent {
  socketId?: string;
  userId?: string;
  role?: MeetRole;
}

/** One remote participant: their own RTCPeerConnection and the stream it delivers. */
export interface MeetPeer {
  id: string; // remote socket id
  userId: string | null;
  role: MeetRole | null;
  stream: MediaStream | null;
  connectionState: RTCPeerConnectionState;
}

interface PeerEntry {
  pc: RTCPeerConnection;
  pendingIce: RTCIceCandidateInit[];
}

const ICE_SERVERS: RTCIceServer[] = [
  { urls: "stun:stun.l.google.com:19302" },
  { urls: "stun:stun1.l.google.com:19302" },
];

/** Older chatroom-service builds don't send socket ids; all signals then share one peer. */
const FALLBACK_PEER_ID = "peer";

/**
 * Video call for a paid appointment, joined through the meeting link
 * booking-service generates (…/meet/<meetingToken>). chatroom-service's
 * meet.socket.controller.js only authorizes the participants and relays
 * `meet:signal`; media flows directly between the browsers.
 *
 * Mesh topology (like Google Meet's small calls): every remote participant gets
 * its own RTCPeerConnection, the local tracks are forwarded into each of them
 * (`addTrack`), and each connection's `ontrack` yields that participant's stream.
 * Whoever is already in the room makes the offer when someone joins
 * (meet:participant-joined), addressing it to the newcomer's socket. A user who
 * is alone stays in the room with their own camera and microphone live.
 */
export function useMeetRoom(meetingToken: string) {
  const socket = useMemo(() => getChatSocket(), []);

  // preparing | joining | waiting | ended | error; connecting/connected are derived from the peers.
  const [baseStatus, setBaseStatus] = useState<MeetStatus>("preparing");
  const [error, setError] = useState<MeetError | null>(null);
  const [role, setRole] = useState<MeetRole | null>(null);
  const [startTime, setStartTime] = useState<number | null>(null);
  const [endTime, setEndTime] = useState<number | null>(null);
  const [presence, setPresence] = useState<MeetPresence | null>(null);
  const [canComplete, setCanComplete] = useState(false);
  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [peers, setPeers] = useState<MeetPeer[]>([]);
  const [isMuted, setIsMuted] = useState(false);
  const [isVideoOff, setIsVideoOff] = useState(false);
  const [isEnding, setIsEnding] = useState(false);

  const localStreamRef = useRef<MediaStream | null>(null);
  const peersRef = useRef<Map<string, PeerEntry>>(new Map());
  const joinedRef = useRef(false);
  const endedRef = useRef(false);

  const patchPeer = useCallback((id: string, patch: Partial<MeetPeer>) => {
    setPeers((list) => list.map((p) => (p.id === id ? { ...p, ...patch } : p)));
  }, []);

  const closePeer = useCallback((id: string) => {
    const entry = peersRef.current.get(id);
    peersRef.current.delete(id);
    if (entry) {
      entry.pc.onicecandidate = null;
      entry.pc.ontrack = null;
      entry.pc.onconnectionstatechange = null;
      entry.pc.close();
    }
    setPeers((list) => list.filter((p) => p.id !== id));
  }, []);

  const closeAllPeers = useCallback(() => {
    [...peersRef.current.keys()].forEach(closePeer);
    setPeers([]);
  }, [closePeer]);

  /** New RTCPeerConnection for one remote participant, carrying our local tracks. */
  const createPeer = useCallback(
    (id: string, info: { userId?: string; role?: MeetRole }) => {
      closePeer(id);
      const pc = new RTCPeerConnection({ iceServers: ICE_SERVERS });
      peersRef.current.set(id, { pc, pendingIce: [] });
      setPeers((list) => [
        ...list,
        { id, userId: info.userId ?? null, role: info.role ?? null, stream: null, connectionState: pc.connectionState },
      ]);

      const local = localStreamRef.current;
      local?.getTracks().forEach((track) => pc.addTrack(track, local));

      const to = id === FALLBACK_PEER_ID ? undefined : id;
      pc.onicecandidate = (event) => {
        if (event.candidate) socket.emit("meet:signal", { to, type: "ice", payload: event.candidate.toJSON() });
      };
      pc.ontrack = (event) => {
        // event.streams[0] is the remote participant's stream for this track.
        const [stream] = event.streams;
        if (stream) patchPeer(id, { stream });
      };
      pc.onconnectionstatechange = () => {
        if (peersRef.current.get(id)?.pc !== pc || endedRef.current) return;
        patchPeer(id, { connectionState: pc.connectionState });
      };
      return pc;
    },
    [socket, closePeer, patchPeer]
  );

  const flushPendingIce = useCallback(async (id: string) => {
    const entry = peersRef.current.get(id);
    if (!entry) return;
    const queued = entry.pendingIce;
    entry.pendingIce = [];
    for (const candidate of queued) {
      try {
        await entry.pc.addIceCandidate(candidate);
      } catch (err) {
        console.warn("meet: dropped ICE candidate", err);
      }
    }
  }, []);

  const makeOffer = useCallback(
    async (id: string, info: { userId?: string; role?: MeetRole }) => {
      const pc = createPeer(id, info);
      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      socket.emit("meet:signal", { to: id === FALLBACK_PEER_ID ? undefined : id, type: "offer", payload: offer });
    },
    [createPeer, socket]
  );

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
        setBaseStatus("error");
        return;
      }
      // Microphone starts live, even while alone in the room.
      stream.getAudioTracks().forEach((t) => (t.enabled = true));
      localStreamRef.current = stream;
      setLocalStream(stream);
      setBaseStatus("joining");
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
      setBaseStatus("joining");
      socket.emit("meet:join", { meetingToken }, (ack: JoinAck) => {
        if (!ack?.ok) {
          joinedRef.current = false;
          setError({ code: ack?.code ?? "unavailable", message: ack?.message ?? "Could not join the meeting." });
          setBaseStatus(ack?.code === "ended" ? "ended" : "error");
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
        // In the room — participants already here will send us their offers.
        setBaseStatus("waiting");
      });
    };

    const onConnect = () => join();
    const onDisconnect = () => {
      joinedRef.current = false;
      closeAllPeers();
      if (!endedRef.current) setBaseStatus("joining");
    };

    const onParticipantJoined = (data: ParticipantEvent) => {
      const id = data?.socketId ?? FALLBACK_PEER_ID;
      makeOffer(id, { userId: data?.userId, role: data?.role }).catch((err) => console.error("meet: offer failed", err));
    };
    const onParticipantLeft = (data: ParticipantEvent) => {
      if (data?.socketId) closePeer(data.socketId);
      else closeAllPeers();
    };

    const onSignal = async (msg: SignalMessage) => {
      const id = msg.fromSocketId ?? FALLBACK_PEER_ID;
      try {
        if (msg.type === "offer") {
          const pc = createPeer(id, { userId: msg.from, role: msg.role });
          await pc.setRemoteDescription(msg.payload as RTCSessionDescriptionInit);
          await flushPendingIce(id);
          const answer = await pc.createAnswer();
          await pc.setLocalDescription(answer);
          socket.emit("meet:signal", { to: msg.fromSocketId, type: "answer", payload: answer });
        } else if (msg.type === "answer") {
          const pc = peersRef.current.get(id)?.pc;
          if (!pc || pc.signalingState !== "have-local-offer") return;
          await pc.setRemoteDescription(msg.payload as RTCSessionDescriptionInit);
          await flushPendingIce(id);
        } else if (msg.type === "ice") {
          const entry = peersRef.current.get(id);
          const candidate = msg.payload as RTCIceCandidateInit;
          if (!entry) return;
          if (entry.pc.remoteDescription) await entry.pc.addIceCandidate(candidate);
          else entry.pendingIce.push(candidate);
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
      closeAllPeers();
      setBaseStatus("ended");
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
      closeAllPeers();
    };
  }, [mediaReady, meetingToken, socket, closePeer, closeAllPeers, createPeer, flushPendingIce, makeOffer]);

  const status: MeetStatus = useMemo(() => {
    if (baseStatus !== "waiting" || peers.length === 0) return baseStatus;
    return peers.some((p) => p.connectionState === "connected") ? "connected" : "connecting";
  }, [baseStatus, peers]);

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
    closeAllPeers();
    localStreamRef.current?.getTracks().forEach((t) => t.stop());
  }, [socket, closeAllPeers]);

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
    peers,
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
