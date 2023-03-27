import React from "react";

export default function VideoPlayer({
  stream,
}: {
  stream: undefined | MediaStream;
}) {
  const videoElement = React.useRef<null | HTMLVideoElement>(null);

  React.useEffect(() => {
    if (videoElement.current === null) {
      return;
    }
    videoElement.current.srcObject = stream || null;
  }, [stream]);

  const show = stream !== undefined;

  /* eslint-disable jsx-a11y/media-has-caption */
  // Lack of closed caption support is a longstanding deficiency in Hypcast.
  // After experimenting with several approaches in GStreamer, I'm ashamed to
  // say that I have yet to identify one that works reliably and consistently.
  // Also, it is not clear that the eventual implementation of closed captions
  // will involve WebVTT, which is what this rule actually looks for (through
  // the presence of a <track> element), so it may need to remain disabled even
  // after closed caption support is in place.
  return (
    <main className="VideoPlayer">
      <video
        style={{ display: show ? undefined : "none" }}
        ref={videoElement}
        autoPlay
        controls
      />
    </main>
  );
}
