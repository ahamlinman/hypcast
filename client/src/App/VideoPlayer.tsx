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
    <div className="VideoPlayer">
      <video ref={videoElement} autoPlay controls />
    </div>
  );
}
