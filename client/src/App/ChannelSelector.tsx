import React, { FormEvent } from "react";

import useConfig from "../useConfig";

export default function ChannelSelector({
  onTune,
}: {
  onTune: (ch: string) => Promise<void>;
}) {
  const channelNames = useConfig<string[]>("channels");

  const [selected, setSelected] = React.useState<undefined | string>();
  const [forceDisabled, setForceDisabled] = React.useState(false);

  React.useEffect(() => {
    if (channelNames instanceof Error) {
      console.error(channelNames);
    }
    if (channelNames instanceof Array) {
      setSelected((s) => (s === undefined ? channelNames[0] : s));
    }
  }, [channelNames]);

  const disabled =
    forceDisabled ||
    channelNames === undefined ||
    channelNames instanceof Error;

  const handleTune = async (evt: FormEvent) => {
    evt.preventDefault();

    if (selected === undefined) {
      throw new Error("tried to tune before channels loaded");
    }

    setForceDisabled(true);
    try {
      await onTune(selected);
    } catch (e) {
      console.error("Tune request failed", e);
    } finally {
      setForceDisabled(false);
    }
  };

  return (
    <p>
      <select
        name="channel"
        value={selected}
        onChange={(evt) => setSelected(evt.currentTarget.value)}
      >
        {channelNames instanceof Array
          ? channelNames.map((ch) => (
              <option key={ch} value={ch}>
                {ch}
              </option>
            ))
          : null}
      </select>
      <button disabled={disabled} onClick={handleTune}>
        Tune
      </button>
    </p>
  );
}
