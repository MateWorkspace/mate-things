const textEncoder = new TextEncoder();
const textDecoder = new TextDecoder();

export function encodeUtf8Json(value: unknown): Uint8Array {
  return textEncoder.encode(JSON.stringify(value));
}

export function decodeUtf8Text(view: DataView): string {
  return textDecoder.decode(view);
}

export function decodeUtf8Json<T>(view: DataView): T {
  return JSON.parse(decodeUtf8Text(view)) as T;
}

export function encodeUint32LE(value: number): Uint8Array {
  const bytes = new Uint8Array(4);
  new DataView(bytes.buffer).setUint32(0, value, true);
  return bytes;
}

export function decodeUint32LE(view: DataView): number {
  return view.getUint32(0, true);
}

export function encodeBoolByte(value: boolean): Uint8Array {
  return new Uint8Array([value ? 1 : 0]);
}

export function decodeBoolByte(view: DataView): boolean {
  return view.getUint8(0) !== 0;
}
