"use client";

import { Mic, Square, X } from "lucide-react";
import { useVoiceRecorder } from "@/hooks";

export interface VoiceRecorderButtonProps {
  onSend: (blob: Blob, duration: number) => void;
}

function formatTime(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export function VoiceRecorderButton({ onSend }: VoiceRecorderButtonProps) {
  const { isRecording, recordingTime, startRecording, stopRecording, cancelRecording } = useVoiceRecorder();

  if (isRecording) {
    return (
      <div className="flex shrink-0 items-center gap-2 px-1">
        <span className="flex items-center gap-1.5 text-xs font-medium text-danger">
          <span className="h-2 w-2 animate-pulse rounded-full bg-danger" />
          {formatTime(recordingTime)}
        </span>
        <button
          type="button"
          onClick={cancelRecording}
          className="inline-flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-surface hover:text-foreground"
          aria-label="Cancel recording"
        >
          <X className="h-4 w-4" />
        </button>
        <button
          type="button"
          onClick={async () => {
            const blob = await stopRecording();
            if (blob) onSend(blob, recordingTime);
          }}
          className="inline-flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-white transition-colors hover:bg-primary-hover"
          aria-label="Send voice message"
        >
          <Square className="h-4 w-4" />
        </button>
      </div>
    );
  }

  return (
    <button
      type="button"
      onClick={startRecording}
      className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-surface hover:text-foreground"
      aria-label="Record voice message"
    >
      <Mic className="h-5 w-5" />
    </button>
  );
}
