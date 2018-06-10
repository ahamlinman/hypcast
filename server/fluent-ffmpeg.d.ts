declare module 'fluent-ffmpeg';

// TODO: The documentation and type definitions for fluent-ffmpeg are
// incorrect. Both claim that the "logger" property given to the constructor
// must have a "warning" method. However, the relevant function of "console" is
// called "console.warn", and I've confirmed that this is *actually* the
// function the library calls.

// Even after fixing this, I think there might be other issues with the typings
// and/or my usage of the library. The real solution is to contribute back to
// fluent-ffmpeg and DefinitelyTyped, but note that the full set of fixes may
// be non-trivial.
