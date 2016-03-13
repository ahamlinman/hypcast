class AzapError extends Error {
  constructor(message, output) {
    super(message);
    this.output = output;
  }
}

export default AzapError;
