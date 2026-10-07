# Forgetting: storage compression before history deletion

## Proposed invariant

Hash the canonical, uncompressed record bytes. Compression is only a storage
representation, so changing codecs must not change record identities.

Skyra already hashes the record encoding when committing and verifies the bytes
when reading. Compression should preserve the **exact committed bytes**: do not
decode JSON and reserialize it while packing records. Inputs likewise retain
their identities based on the exact observed bytes.

## Suggested approach

Start with sealed segments and Zstd, measure savings on actual Skyra logs, and
defer history deletion until replay and recovery requirements are defined. There
is no need to invent a new compression algorithm first.

1. Keep active writes in the existing format. Pack completed runs into immutable
   segments first.
2. Use independently compressed Zstd chunks with an index mapping object hashes
   to chunks and offsets. Reading one record should not require decompressing an
   entire archive. The [Zstd seekable format](https://github.com/facebook/zstd/blob/dev/contrib/seekable_format/zstd_seekable_compression_format.md)
   provides a reference for independently compressed frames and indexed access.
3. Make storage readers return the original bytes and verify their hashes after
   decompression. Codec choice must remain invisible to record identity and replay.
4. Preserve record timestamps. The chat-log reader currently uses them for
   ordering records and calculating tool-call durations.
5. Measure actual disk savings, individual object lookup latency, and full replay
   time on real conversations. Verify replay and recovery against packed storage
   before relying on it.

## Compaction and forgetting are separate decisions

Removing loose files after a verified archive durably replaces them is storage
compaction: the same history remains available in another representation.

Deleting history from every representation is a retention policy. Defer that
decision until the history required by replay and recovery is explicitly defined.

The main engineering work is making every reader, including chat history and
recovery, understand packed storage. Zstd itself is the straightforward part.
