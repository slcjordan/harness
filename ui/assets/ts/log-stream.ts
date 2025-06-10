export function logStream(url: string) {
    return {
          logOutput: '',
              ws: null as WebSocket | null,
                  start() {
                          this.ws = new WebSocket(url);

                                this.ws.onmessage = (event: MessageEvent) => {
                                          this.logOutput += event.data + '\n';
                                                };

                                                      this.ws.onclose = () => {
                                                                this.logOutput += '[Disconnected]\n';
                                                                      };

                                                                            this.ws.onerror = () => {
                                                                                      this.logOutput += '[Error]\n';
                                                                                            };
                                                                                                }
                                                                                                  };
}

