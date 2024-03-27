import React from "react";

import useConfig from "../useConfig";

export default function ChannelSelector({
  selected,
  onTune,
}: {
  selected?: string;
  onTune: (ch: string) => Promise<void>;
}) {
  const channelNames = useConfig<string[]>("channels");

  return channelNames instanceof Array ? (
    <aside className="ChannelSelector">
      {channelNames.map((ch) => (
        <Channel
          key={ch}
          name={ch}
          active={ch === selected}
          onClick={() => {
            onTune(ch).catch((e) => console.error("Tune request failed", e));
          }}
        />
      ))}
    </aside>
  ) : null;
}

function Channel({
  name,
  active,
  onClick,
}: {
  name: string;
  active?: boolean;
  onClick: () => void;
}) {
  return (
    <button
      className={`ChannelSelector__Channel ${
        active ? "ChannelSelector__Channel--Active" : ""
      }`}
      onClick={onClick}
    >
      {name}
    </button>
  );
}
