declare class Argon2Result {
  encoded: string;
  hash: Uint8Array;
  hashHex: string;
}

declare var argon2: {
  type: 2,
  hash: (arg: {
    pass: string,
    salt: string,
    hashLen: number,
    type: number,
    mem: number,
    time: number,
    parallelism: number,
  }) => Promise<Argon2Result>;
};
