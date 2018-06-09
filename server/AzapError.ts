export default class AzapError extends Error {
  constructor(message, public readonly output) {
    super(message);
  }
}
