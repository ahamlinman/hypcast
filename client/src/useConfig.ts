import React from "react";

export default function useConfig<T>(name: string): undefined | T | Error {
  const [result, setResult] = React.useState<undefined | T | Error>();

  React.useEffect(() => {
    setResult(undefined);
    startFetch();

    async function startFetch() {
      try {
        const response = await fetch(`/api/config/${name}`);
        if (!response.ok) {
          throw new Error(
            `failed to retrieve ${name} config: ${response.statusText}`,
          );
        }
        setResult(await response.json());
      } catch (e) {
        setResult(e);
      }
    }
  }, [name]);

  return result;
}
