/**
 * Jest mock for @aws-sdk/client-bedrock-runtime.
 */
export class BedrockRuntimeClient {
  constructor(_config: any) {}
  send(_command: any) { return Promise.resolve({ output: { message: { content: [{ text: 'Mock response' }] } } }); }
}

export class ConverseCommand {
  constructor(_input: any) {}
}
