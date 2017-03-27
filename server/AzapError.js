export default class AzapError extends Error {
  constructor(message, output) {
    super(message);
    this.output = output;
  }
}
