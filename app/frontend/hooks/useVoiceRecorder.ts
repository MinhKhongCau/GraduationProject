"use client";

import { useCallback, useRef, useState } from "react";

const MIME_CANDIDATES: Array<{ mimeType: string; bitrate: number }> = [
  { mimeType: "audio/webm;codecs=opus", bitrate: 256_000 },
  { mimeType: "audio/webm", bitrate: 192_000 },
  { mimeType: "audio/ogg;codecs=opus", bitrate: 192_000 },
  { mimeType: "audio/mp4", bitrate: 128_000 },
];

function pickMimeType(): { mimeType: string; bitrate: number } {
  for (const candidate of MIME_CANDIDATES) {
    if (MediaRecorder.isTypeSupported(candidate.mimeType)) return candidate;
  }
  return { mimeType: "", bitrate: 128_000 };
}

export interface UseVoiceRecorder {
  isRecording: boolean;
  recordingTime: number;
  startRecording: () => Promise<void>;
  stopRecording: () => Promise<Blob | null>;
  cancelRecording: () => void;
}

/** Push-to-record voice-message hook, ported from the reference client's
 * useVoiceRecorder.js (MediaRecorder, 50ms timeslice, MIME negotiation). */
export function useVoiceRecorder(): UseVoiceRecorder {
  const [isRecording, setIsRecording] = useState(false);
  const [recordingTime, setRecordingTime] = useState(0);

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const streamRef = useRef<MediaStream | null>(null);

  const clearTimer = () => {
    if (timerRef.current) clearInterval(timerRef.current);
    timerRef.current = null;
  };

  const stopTracks = () => {
    streamRef.current?.getTracks().forEach((track) => track.stop());
    streamRef.current = null;
  };

  const startRecording = useCallback(async () => {
    let stream: MediaStream;
    try {
      stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true,
          sampleRate: 48000,
          channelCount: 1,
        },
      });
    } catch {
      alert("Microphone access denied. Please allow microphone permission and try again.");
      return;
    }

    streamRef.current = stream;
    const { mimeType, bitrate } = pickMimeType();
    const recorder = new MediaRecorder(stream, {
      ...(mimeType ? { mimeType } : {}),
      audioBitsPerSecond: bitrate,
    });

    chunksRef.current = [];
    recorder.ondataavailable = (event) => {
      if (event.data.size > 0) chunksRef.current.push(event.data);
    };

    mediaRecorderRef.current = recorder;
    recorder.start(50);
    setIsRecording(true);
    setRecordingTime(0);
    timerRef.current = setInterval(() => setRecordingTime((t) => t + 1), 1000);
  }, []);

  const stopRecording = useCallback((): Promise<Blob | null> => {
    const recorder = mediaRecorderRef.current;
    if (!recorder || !streamRef.current) return Promise.resolve(null);

    return new Promise((resolve) => {
      recorder.onstop = () => {
        const blob = new Blob(chunksRef.current, { type: recorder.mimeType });
        stopTracks();
        chunksRef.current = [];
        resolve(blob);
      };
      recorder.stop();
      setIsRecording(false);
      clearTimer();
      setRecordingTime(0);
    });
  }, []);

  const cancelRecording = useCallback(() => {
    mediaRecorderRef.current?.stop();
    stopTracks();
    chunksRef.current = [];
    setIsRecording(false);
    clearTimer();
    setRecordingTime(0);
  }, []);

  return { isRecording, recordingTime, startRecording, stopRecording, cancelRecording };
}
