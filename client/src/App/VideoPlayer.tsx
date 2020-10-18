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

  return (
    <video
      style={{ border: "1px solid black" }}
      ref={videoElement}
      autoPlay
      controls
    />
  );
}
