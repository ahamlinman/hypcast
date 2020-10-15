const NO_CONTENT = 204;

export default async function rpc(name: string, params?: { [k: string]: any }) {
  const response = await fetch(`/api/rpc/${name}`, {
    method: "POST",
    body: params ? JSON.stringify(params) : undefined,
    headers: params
      ? new Headers({ "Content-Type": "application/json" })
      : undefined,
  });

  if (!response.ok) {
    const body = await response.json();
    throw new Error(body.Error);
  }

  if (response.status !== NO_CONTENT) {
    return await response.json();
  }
}
