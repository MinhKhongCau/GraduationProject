"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { getChatSocket } from "@/lib/chatSocket";
import { blobToBase64 } from "@/lib/audio";
import { useAuthContext } from "./AuthContext";
import type {
  ChatConnectionState,
  ChatMessage,
  DmHistoryPayload,
  DmReactionPayload,
  DmTypingStatusPayload,
  MessageReactions,
} from "@/types";
import { CallOverlay } from "@/components/chat/CallOverlay";

const TYPING_TIMEOUT_MS = 3000;

interface CallState {
  status: "idle" | "calling" | "incoming" | "connecting" | "connected";
  isCaller: boolean;
  peerId: string | null;
  peerName?: string;
  peerAvatar?: string;
}

interface ChatContextValue {
  connectionState: ChatConnectionState;
  messagesByDmId: Record<string, ChatMessage[]>;
  reactionsByMessageId: MessageReactions;
  typingByContactId: Record<string, boolean>;
  onlineByContactId: Record<string, boolean>;
  fetchHistory: (contactId: string, page?: number, limit?: number, lastMessageId?: string | null) => void;
  sendText: (contactId: string, text: string) => void;
  sendVoice: (contactId: string, blob: Blob, duration: number) => Promise<void>;
  react: (contactId: string, messageId: string, emoji: string) => void;
  setTyping: (contactId: string, isTyping: boolean) => void;
  queryPresence: (contactIds: string[]) => void;
  // Call functionality
  callState: CallState;
  localStream: MediaStream | null;
  remoteStream: MediaStream | null;
  isMuted: boolean;
  isVideoOff: boolean;
  startCall: (contactId: string, contactName: string, contactAvatar?: string) => void;
  acceptCall: () => void;
  rejectCall: () => void;
  endCall: () => void;
  toggleMute: () => void;
  toggleVideo: () => void;
}

const ChatContext = createContext<ChatContextValue | null>(null);

export function ChatProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated, user } = useAuthContext();
  const socket = useMemo(() => getChatSocket(), []);

  const [connectionState, setConnectionState] = useState<ChatConnectionState>("idle");
  const [messagesByDmId, setMessagesByDmId] = useState<Record<string, ChatMessage[]>>({});
  const [reactionsByMessageId, setReactionsByMessageId] = useState<MessageReactions>({});
  const [typingByContactId, setTypingByContactId] = useState<Record<string, boolean>>({});
  const [onlineByContactId, setOnlineByContactId] = useState<Record<string, boolean>>({});

  // Call states
  const [callState, setCallState] = useState<CallState>({
    status: "idle",
    isCaller: false,
    peerId: null,
  });
  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStream, setRemoteStream] = useState<MediaStream | null>(null);
  const [isMuted, setIsMuted] = useState(false);
  const [isVideoOff, setIsVideoOff] = useState(false);

  const peerConnectionRef = useRef<RTCPeerConnection | null>(null);
  const localStreamRef = useRef<MediaStream | null>(null);
  const typingTimeoutsRef = useRef<Record<string, ReturnType<typeof setTimeout>>>({});

  // Register online status for calls when connected
  useEffect(() => {
    if (connectionState === "connected" && user?.id) {
      socket.emit("user:online", { name: user.id });
    }
  }, [connectionState, user, socket]);

  const cleanupCall = useCallback(() => {
    if (peerConnectionRef.current) {
      peerConnectionRef.current.close();
      peerConnectionRef.current = null;
    }
    if (localStreamRef.current) {
      localStreamRef.current.getTracks().forEach((track) => track.stop());
      localStreamRef.current = null;
    }
    setLocalStream(null);
    setRemoteStream(null);
    setCallState({ status: "idle", isCaller: false, peerId: null });
    setIsMuted(false);
    setIsVideoOff(false);
  }, []);

  const setupPeerConnection = useCallback(
    async (targetUserId: string, isInitiator: boolean) => {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          video: true,
          audio: true,
        });
        localStreamRef.current = stream;
        setLocalStream(stream);

        const pc = new RTCPeerConnection({
          iceServers: [
            { urls: "stun:stun.l.google.com:19302" },
            { urls: "stun:stun1.l.google.com:19302" },
          ],
        });
        peerConnectionRef.current = pc;

        stream.getTracks().forEach((track) => {
          pc.addTrack(track, stream);
        });

        pc.onicecandidate = (event) => {
          if (event.candidate) {
            socket.emit("webrtc:ice", { to: targetUserId, candidate: event.candidate });
          }
        };

        pc.ontrack = (event) => {
          if (event.streams && event.streams[0]) {
            setRemoteStream(event.streams[0]);
            setCallState((prev) => ({ ...prev, status: "connected" }));
          }
        };

        if (isInitiator) {
          const offer = await pc.createOffer();
          await pc.setLocalDescription(offer);
          socket.emit("webrtc:offer", { to: targetUserId, offer });
        }

        return pc;
      } catch (err) {
        console.error("Failed to setup WebRTC peer connection:", err);
        socket.emit("call:end", { to: targetUserId });
        cleanupCall();
      }
    },
    [socket, cleanupCall]
  );

  useEffect(() => {
    if (!isAuthenticated) {
      socket.disconnect();
      setConnectionState("idle");
      return;
    }

    setConnectionState("connecting");
    socket.connect();

    const onConnect = () => {
      setConnectionState("connected");
      if (user?.id) {
        socket.emit("user:online", { name: user.id });
      }
    };
    const onDisconnect = () => setConnectionState("idle");
    const onConnectError = () => setConnectionState("error");

    const onDmMessage = (msg: ChatMessage) => {
      setMessagesByDmId((prev) => {
        const existing = prev[msg.dmId] ?? [];
        if (existing.some((m) => m.id === msg.id)) return prev;
        return { ...prev, [msg.dmId]: [...existing, msg] };
      });
    };

    const onDmHistory = ({ dmId, history, page = 1 }: DmHistoryPayload) => {
      setMessagesByDmId((prev) => {
        const existing = prev[dmId] ?? [];
        if (page === 1) {
          return { ...prev, [dmId]: history };
        } else {
          const existingIds = new Set(existing.map((m) => m.id));
          const filteredHistory = history.filter((m) => !existingIds.has(m.id));
          return { ...prev, [dmId]: [...filteredHistory, ...existing] };
        }
      });
    };

    const onDmReaction = ({ messageId, emoji, userId }: DmReactionPayload) => {
      setReactionsByMessageId((prev) => {
        const forMessage = { ...(prev[messageId] ?? {}) };
        const users = forMessage[emoji] ?? [];
        forMessage[emoji] = users.includes(userId)
          ? users.filter((id) => id !== userId)
          : [...users, userId];
        if (forMessage[emoji].length === 0) delete forMessage[emoji];
        return { ...prev, [messageId]: forMessage };
      });
    };

    const onDmTypingStatus = ({ fromId, isTyping }: DmTypingStatusPayload) => {
      setTypingByContactId((prev) => ({ ...prev, [fromId]: isTyping }));
      const timeouts = typingTimeoutsRef.current;
      if (timeouts[fromId]) clearTimeout(timeouts[fromId]);
      if (isTyping) {
        timeouts[fromId] = setTimeout(() => {
          setTypingByContactId((prev) => ({ ...prev, [fromId]: false }));
        }, TYPING_TIMEOUT_MS);
      }
    };

    const onPresenceStatus = (status: Record<string, boolean>) => {
      setOnlineByContactId((prev) => ({ ...prev, ...status }));
    };

    // Call signaling events
    const onCallIncoming = async ({ from }: { from: string }) => {
      let name = "Consultant / Patient";
      let avatar = "";
      try {
        if (user?.role === "PATIENT") {
          const { getExpertProfile } = await import("@/api/expert");
          const profile = await getExpertProfile(from);
          name = profile.fullName;
          avatar = profile.avatarUrl || "";
        } else if (user?.role === "EXPERT") {
          const { getPatientProfile } = await import("@/api/patient");
          const profile = await getPatientProfile(from);
          name = profile.fullName;
          avatar = profile.avatarUrl || "";
        }
      } catch (e) {
        console.error("Failed to load caller profile:", e);
      }

      setCallState({
        status: "incoming",
        isCaller: false,
        peerId: from,
        peerName: name,
        peerAvatar: avatar,
      });
    };

    const onCallAccepted = async ({ from }: { from: string }) => {
      setCallState((prev) => ({ ...prev, status: "connecting", peerId: from }));
      await setupPeerConnection(from, true);
    };

    const onCallRejected = () => {
      alert("Call was declined.");
      cleanupCall();
    };

    const onCallUnavailable = () => {
      alert("User is currently offline or unavailable.");
      cleanupCall();
    };

    const onCallEnded = () => {
      cleanupCall();
    };

    const onWebRtcOffer = async ({ from, offer }: { from: string; offer: any }) => {
      const pc = peerConnectionRef.current;
      if (pc) {
        try {
          await pc.setRemoteDescription(new RTCSessionDescription(offer));
          const answer = await pc.createAnswer();
          await pc.setLocalDescription(answer);
          socket.emit("webrtc:answer", { to: from, answer });
        } catch (err) {
          console.error("Error setting up offer:", err);
        }
      }
    };

    const onWebRtcAnswer = async ({ answer }: { from: string; answer: any }) => {
      const pc = peerConnectionRef.current;
      if (pc) {
        try {
          await pc.setRemoteDescription(new RTCSessionDescription(answer));
        } catch (err) {
          console.error("Error setting remote answer:", err);
        }
      }
    };

    const onWebRtcIce = async ({ candidate }: { from: string; candidate: any }) => {
      const pc = peerConnectionRef.current;
      if (pc) {
        try {
          await pc.addIceCandidate(new RTCIceCandidate(candidate));
        } catch (err) {
          console.warn("Error adding ICE candidate:", err);
        }
      }
    };

    socket.on("connect", onConnect);
    socket.on("disconnect", onDisconnect);
    socket.on("connect_error", onConnectError);
    socket.on("dm:message", onDmMessage);
    socket.on("dm:history", onDmHistory);
    socket.on("dm:reaction", onDmReaction);
    socket.on("dm:typing:status", onDmTypingStatus);
    socket.on("presence:status", onPresenceStatus);

    socket.on("call:incoming", onCallIncoming);
    socket.on("call:accepted", onCallAccepted);
    socket.on("call:rejected", onCallRejected);
    socket.on("call:unavailable", onCallUnavailable);
    socket.on("call:ended", onCallEnded);
    socket.on("webrtc:offer", onWebRtcOffer);
    socket.on("webrtc:answer", onWebRtcAnswer);
    socket.on("webrtc:ice", onWebRtcIce);

    return () => {
      socket.off("connect", onConnect);
      socket.off("disconnect", onDisconnect);
      socket.off("connect_error", onConnectError);
      socket.off("dm:message", onDmMessage);
      socket.off("dm:history", onDmHistory);
      socket.off("dm:reaction", onDmReaction);
      socket.off("dm:typing:status", onDmTypingStatus);
      socket.off("presence:status", onPresenceStatus);

      socket.off("call:incoming", onCallIncoming);
      socket.off("call:accepted", onCallAccepted);
      socket.off("call:rejected", onCallRejected);
      socket.off("call:unavailable", onCallUnavailable);
      socket.off("call:ended", onCallEnded);
      socket.off("webrtc:offer", onWebRtcOffer);
      socket.off("webrtc:answer", onWebRtcAnswer);
      socket.off("webrtc:ice", onWebRtcIce);
    };
  }, [isAuthenticated, socket, user, setupPeerConnection, cleanupCall]);

  const fetchHistory = useCallback(
    (contactId: string, page = 1, limit = 20, lastMessageId: string | null = null) =>
      socket.emit("dm:history", { toUser: contactId, page, limit, lastMessageId }),
    [socket]
  );

  const sendText = useCallback(
    (contactId: string, text: string) => {
      const trimmed = text.trim();
      if (!trimmed) return;
      socket.emit("dm:send", { toUser: contactId, text: trimmed });
    },
    [socket]
  );

  const sendVoice = useCallback(
    async (contactId: string, blob: Blob, duration: number) => {
      const audio = await blobToBase64(blob);
      socket.emit("dm:send:voice", { toUser: contactId, audio, duration, mimeType: blob.type });
    },
    [socket]
  );

  const react = useCallback(
    (contactId: string, messageId: string, emoji: string) =>
      socket.emit("dm:react", { toUser: contactId, messageId, emoji }),
    [socket]
  );

  const setTyping = useCallback(
    (contactId: string, isTyping: boolean) =>
      socket.emit("dm:typing", { toUser: contactId, isTyping }),
    [socket]
  );

  const queryPresence = useCallback(
    (contactIds: string[]) => {
      if (contactIds.length > 0) socket.emit("presence:query", { userIds: contactIds });
    },
    [socket]
  );

  // Call actions
  const startCall = useCallback(
    (contactId: string, contactName: string, contactAvatar?: string) => {
      if (!user?.id) return;
      // Re-register call status
      socket.emit("user:online", { name: user.id });

      setCallState({
        status: "calling",
        isCaller: true,
        peerId: contactId,
        peerName: contactName,
        peerAvatar: contactAvatar,
      });

      socket.emit("call:request", { to: contactId });
    },
    [socket, user]
  );

  const acceptCall = useCallback(async () => {
    const peerId = callState.peerId;
    if (!peerId) return;

    setCallState((prev) => ({ ...prev, status: "connecting" }));
    const pc = await setupPeerConnection(peerId, false);
    if (pc) {
      socket.emit("call:accept", { to: peerId });
    }
  }, [socket, callState.peerId, setupPeerConnection]);

  const rejectCall = useCallback(() => {
    const peerId = callState.peerId;
    if (peerId) {
      socket.emit("call:reject", { to: peerId });
    }
    cleanupCall();
  }, [socket, callState.peerId, cleanupCall]);

  const endCall = useCallback(() => {
    const peerId = callState.peerId;
    if (peerId) {
      socket.emit("call:end", { to: peerId });
    }
    cleanupCall();
  }, [socket, callState.peerId, cleanupCall]);

  const toggleMute = useCallback(() => {
    if (localStreamRef.current) {
      const audioTracks = localStreamRef.current.getAudioTracks();
      audioTracks.forEach((track) => {
        track.enabled = !track.enabled;
      });
      setIsMuted((prev) => !prev);
    }
  }, []);

  const toggleVideo = useCallback(() => {
    if (localStreamRef.current) {
      const videoTracks = localStreamRef.current.getVideoTracks();
      videoTracks.forEach((track) => {
        track.enabled = !track.enabled;
      });
      setIsVideoOff((prev) => !prev);
    }
  }, []);

  const value = useMemo(
    () => ({
      connectionState,
      messagesByDmId,
      reactionsByMessageId,
      typingByContactId,
      onlineByContactId,
      fetchHistory,
      sendText,
      sendVoice,
      react,
      setTyping,
      queryPresence,
      // Call values
      callState,
      localStream,
      remoteStream,
      isMuted,
      isVideoOff,
      startCall,
      acceptCall,
      rejectCall,
      endCall,
      toggleMute,
      toggleVideo,
    }),
    [
      connectionState,
      messagesByDmId,
      reactionsByMessageId,
      typingByContactId,
      onlineByContactId,
      fetchHistory,
      sendText,
      sendVoice,
      react,
      setTyping,
      queryPresence,
      // Call dependencies
      callState,
      localStream,
      remoteStream,
      isMuted,
      isVideoOff,
      startCall,
      acceptCall,
      rejectCall,
      endCall,
      toggleMute,
      toggleVideo,
    ]
  );

  return (
    <ChatContext.Provider value={value}>
      {children}
      <CallOverlay
        status={callState.status}
        isCaller={callState.isCaller}
        peerName={callState.peerName}
        peerAvatar={callState.peerAvatar}
        localStream={localStream}
        remoteStream={remoteStream}
        isMuted={isMuted}
        isVideoOff={isVideoOff}
        acceptCall={acceptCall}
        rejectCall={rejectCall}
        endCall={endCall}
        toggleMute={toggleMute}
        toggleVideo={toggleVideo}
      />
    </ChatContext.Provider>
  );
}

export function useChatContext(): ChatContextValue {
  const context = useContext(ChatContext);
  if (!context) {
    throw new Error("useChatContext must be used within a ChatProvider");
  }
  return context;
}

