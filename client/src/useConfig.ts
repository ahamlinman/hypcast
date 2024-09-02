import React from "react";

export default function useConfig<T>(name: string): undefined | T | Error {
  const [result, setResult] = React.useState<undefined | T | Error>();

  React.useEffect(() => {
    setResult(undefined);
    startFetch();

    async function startFetch() {
      try {
        setResult(await fetchConfigWithCache(name));
      } catch (e: unknown) {
        setResult(e as Error);
      }
    }
  }, [name]);

  return result;
}

const FETCH_CACHE = new Map<string, Promise<unknown>>();

function fetchConfigWithCache(name: string) {
  const cachedPromise = FETCH_CACHE.get(name);
  if (cachedPromise !== undefined) {
    return cachedPromise;
  }

  const responsePromise = fetch(`/api/config/${name}`).then((response) => {
    if (!response.ok) {
      return new Error(
        `failed to retrieve ${name} config: ${response.statusText}`,
      );
    }
    return response.json();
  });

  FETCH_CACHE.set(name, responsePromise);
  return responsePromise;
}
