/** Strips the `data:...;base64,` prefix FileReader adds, matching the raw
 * base64 chatroom-service expects for voice-message payloads. */
export function blobToBase64(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve((reader.result as string).split(",")[1] ?? "");
    reader.onerror = reject;
    reader.readAsDataURL(blob);
  });
}

/** Decodes+plays a base64 voice message with gain amplification (source
 * clips are recorded quietly) — same approach as the reference client. */
export async function playBase64Audio(base64: string, gain = 3.0): Promise<void> {
  const raw = atob(base64);
  const bytes = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i);

  const ctx = new AudioContext();
  const audioBuffer = await ctx.decodeAudioData(bytes.buffer);
  const source = ctx.createBufferSource();
  source.buffer = audioBuffer;

  const gainNode = ctx.createGain();
  gainNode.gain.value = gain;

  source.connect(gainNode).connect(ctx.destination);
  source.start(0);
  source.onended = () => ctx.close();
}
