export default class AzapError extends Error {
  constructor(message: string, public readonly output: string) {
    super(message);
  }
}
