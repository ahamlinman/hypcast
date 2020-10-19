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
