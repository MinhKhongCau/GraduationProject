"use client";

import { useEffect, useRef } from "react";
import { Phone, PhoneOff, Mic, MicOff, Video, VideoOff, User } from "lucide-react";
import Image from "next/image";

interface CallOverlayProps {
  status: "idle" | "calling" | "incoming" | "connecting" | "connected";
  isCaller: boolean;
  peerName?: string;
  peerAvatar?: string;
  localStream: MediaStream | null;
  remoteStream: MediaStream | null;
  isMuted: boolean;
  isVideoOff: boolean;
  acceptCall: () => void;
  rejectCall: () => void;
  endCall: () => void;
  toggleMute: () => void;
  toggleVideo: () => void;
}

export function CallOverlay({
  status,
  isCaller,
  peerName = "User",
  peerAvatar,
  localStream,
  remoteStream,
  isMuted,
  isVideoOff,
  acceptCall,
  rejectCall,
  endCall,
  toggleMute,
  toggleVideo,
}: CallOverlayProps) {
  const localVideoRef = useRef<HTMLVideoElement | null>(null);
  const remoteVideoRef = useRef<HTMLVideoElement | null>(null);

  useEffect(() => {
    if (localVideoRef.current && localStream) {
      localVideoRef.current.srcObject = localStream;
    }
  }, [localStream, status]);

  useEffect(() => {
    if (remoteVideoRef.current && remoteStream) {
      remoteVideoRef.current.srcObject = remoteStream;
    }
  }, [remoteStream, status]);

  if (status === "idle") return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-md p-4 text-white">
      <div className="relative flex h-full max-h-[800px] w-full max-w-[600px] flex-col overflow-hidden rounded-2xl bg-neutral-900 border border-neutral-800 shadow-elevated">
        {/* Active Call UI */}
        {status === "connected" && (
          <div className="relative flex-1 bg-black">
            {/* Remote Video (Full Screen inside container) */}
            <video
              ref={remoteVideoRef}
              autoPlay
              playsInline
              className="h-full w-full object-cover"
            />

            {/* Local Video (Thumbnail overlay) */}
            {!isVideoOff && localStream && (
              <div className="absolute right-4 top-4 h-36 w-28 overflow-hidden rounded-xl border border-neutral-700 bg-neutral-800 shadow-card">
                <video
                  ref={localVideoRef}
                  autoPlay
                  playsInline
                  muted
                  className="h-full w-full object-cover"
                />
              </div>
            )}

            {/* Peer info Overlay */}
            <div className="absolute left-4 top-4 flex items-center gap-2 rounded-full bg-black/40 px-3 py-1.5 backdrop-blur-sm">
              <div className="h-2 w-2 rounded-full bg-success animate-pulse" />
              <span className="text-xs font-semibold">{peerName}</span>
            </div>
          </div>
        )}

        {/* Calling / Incoming / Connecting Screen */}
        {status !== "connected" && (
          <div className="flex flex-1 flex-col items-center justify-center p-8 text-center">
            <div className="relative mb-6">
              {/* Pulsating Ring for call incoming/outgoing */}
              <div className="absolute inset-0 rounded-full bg-primary/20 animate-ping" />
              {peerAvatar ? (
                <Image
                  src={peerAvatar}
                  alt={peerName}
                  width={120}
                  height={120}
                  unoptimized
                  className="relative h-28 w-28 rounded-full border-4 border-neutral-800 object-cover shadow-lg"
                />
              ) : (
                <div className="relative flex h-28 w-28 items-center justify-center rounded-full border-4 border-neutral-800 bg-neutral-800 text-neutral-400">
                  <User className="h-14 w-14" />
                </div>
              )}
            </div>

            <h2 className="text-xl font-bold tracking-tight">{peerName}</h2>
            
            <p role="status" className="mt-2 text-sm text-neutral-300 font-medium">
              {status === "incoming" && "Incoming video call..."}
              {status === "calling" && "Calling..."}
              {status === "connecting" && "Connecting..."}
            </p>
          </div>
        )}

        {/* Control Bar */}
        <div className="flex items-center justify-center gap-6 border-t border-neutral-800 bg-neutral-950/80 p-6 backdrop-blur-md">
          {status === "incoming" ? (
            <>
              {/* Accept & Reject Buttons */}
              <button
                type="button"
                onClick={rejectCall}
                className="flex h-14 w-14 items-center justify-center rounded-full bg-danger text-white transition-colors hover:bg-danger/85"
                title="Decline Call"
                aria-label="Decline Call"
              >
                <PhoneOff className="h-6 w-6" />
              </button>
              <button
                type="button"
                onClick={acceptCall}
                className="flex h-14 w-14 items-center justify-center rounded-full bg-success text-white transition-colors hover:bg-success/85"
                title="Accept Call"
                aria-label="Accept Call"
              >
                <Phone className="h-6 w-6" />
              </button>
            </>
          ) : (
            <>
              {/* Active Call Controls */}
              {status === "connected" && (
                <>
                  <button
                    type="button"
                    onClick={toggleMute}
                    className={`flex h-12 w-12 items-center justify-center rounded-full transition-colors ${
                      isMuted ? "bg-danger hover:bg-danger/85" : "bg-neutral-800 hover:bg-neutral-700"
                    }`}
                    title={isMuted ? "Unmute Mic" : "Mute Mic"}
                    aria-label={isMuted ? "Unmute Mic" : "Mute Mic"}
                    aria-pressed={isMuted}
                  >
                    {isMuted ? <MicOff className="h-5 w-5" /> : <Mic className="h-5 w-5" />}
                  </button>

                  <button
                    type="button"
                    onClick={toggleVideo}
                    className={`flex h-12 w-12 items-center justify-center rounded-full transition-colors ${
                      isVideoOff ? "bg-danger hover:bg-danger/85" : "bg-neutral-800 hover:bg-neutral-700"
                    }`}
                    title={isVideoOff ? "Turn Video On" : "Turn Video Off"}
                    aria-label={isVideoOff ? "Turn Video On" : "Turn Video Off"}
                    aria-pressed={isVideoOff}
                  >
                    {isVideoOff ? <VideoOff className="h-5 w-5" /> : <Video className="h-5 w-5" />}
                  </button>
                </>
              )}

              {/* End / Cancel Call Button */}
              <button
                type="button"
                onClick={endCall}
                className="flex h-14 w-14 items-center justify-center rounded-full bg-danger text-white transition-colors hover:bg-danger/85"
                title="End Call"
                aria-label="End Call"
              >
                <PhoneOff className="h-6 w-6" />
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
