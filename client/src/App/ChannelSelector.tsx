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

  const handleTune = async (selected: string) => {
    if (selected === undefined) {
      throw new Error("tried to tune before channels loaded");
    }

    try {
      await onTune(selected);
    } catch (e) {
      console.error("Tune request failed", e);
    }
  };

  if (!(channelNames instanceof Array)) {
    return null;
  }

  return (
    <div className="ChannelSelector">
      {channelNames.map((ch) => (
        <Channel
          key={ch}
          name={ch}
          active={ch === selected}
          onClick={() => {
            handleTune(ch);
          }}
        />
      ))}
    </div>
  );
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
