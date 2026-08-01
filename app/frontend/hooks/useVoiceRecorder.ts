"use client";

import { useCallback, useRef, useState } from "react";
import { Capacitor } from "@capacitor/core";

const WEB_MIME_CANDIDATES: Array<{ mimeType: string; bitrate: number }> = [
  { mimeType: "audio/webm;codecs=opus", bitrate: 256_000 },
  { mimeType: "audio/webm", bitrate: 192_000 },
  { mimeType: "audio/ogg;codecs=opus", bitrate: 192_000 },
  { mimeType: "audio/mp4", bitrate: 128_000 },
];

// Safari/WKWebView records AAC in an MP4 container when MediaRecorder supports it.
const IOS_MIME_CANDIDATES: Array<{ mimeType: string; bitrate: number }> = [
  { mimeType: "audio/mp4", bitrate: 128_000 },
  ...WEB_MIME_CANDIDATES,
];

function pickMimeType(): { mimeType: string; bitrate: number } {
  const candidates = Capacitor.getPlatform() === "ios" ? IOS_MIME_CANDIDATES : WEB_MIME_CANDIDATES;
  for (const candidate of candidates) {
    if (MediaRecorder.isTypeSupported(candidate.mimeType)) return candidate;
  }
  return { mimeType: "", bitrate: 128_000 };
}

async function ensureMicrophonePermission(): Promise<void> {
  if (!navigator.mediaDevices?.getUserMedia) {
    throw new Error("Voice recording is not supported by this browser.");
  }

  // Capacitor has no generic microphone-permissions plugin in v8. The native
  // WebView requests the declared iOS/Android microphone permission through
  // getUserMedia; query first when the browser exposes the Permissions API.
  try {
    const permission = await navigator.permissions?.query({ name: "microphone" as PermissionName });
    if (permission?.state === "denied") {
      throw new Error(
        Capacitor.isNativePlatform()
          ? "Microphone access is disabled. Enable it for MindCare in your device settings."
          : "Microphone access is blocked. Enable it in your browser settings and try again."
      );
    }
  } catch (error) {
    if (error instanceof Error && error.message.includes("Microphone")) throw error;
    // Some WebViews do not implement navigator.permissions for microphones.
  }
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
      await ensureMicrophonePermission();
      stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true,
          sampleRate: 48000,
          channelCount: 1,
        },
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : "Microphone access was denied. Please allow it and try again.";
      alert(message);
      return;
    }

    streamRef.current = stream;
    if (typeof MediaRecorder === "undefined") {
      stopTracks();
      alert("This device does not support voice recording.");
      return;
    }

    const { mimeType, bitrate } = pickMimeType();
    let recorder: MediaRecorder;
    try {
      recorder = new MediaRecorder(stream, {
        ...(mimeType ? { mimeType } : {}),
        audioBitsPerSecond: bitrate,
      });
    } catch {
      stopTracks();
      alert("This device does not support a compatible audio recording format.");
      return;
    }

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
