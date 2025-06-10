function streamComponent() {
  return {
    connected: false,
    status: "Disconnected",
    data: {},
    socket: null,

    connect(url) {
      const socket = new WebSocket(url);

      socket.onopen = () => {
        this.connected = true;
        this.status = "Connected";
      };

      socket.onmessage = (e) => {
        this.data = { ...this.data, ...json.parse(e.data)};
      };

      socket.onclose = () => {
        this.connected = false;
        this.status = "Disconnected";
      };

      socket.onerror = () => {
        this.connected = false;
        this.status = "Error";
      };

      this.sendMessage = (msg) => {
        if (socket.readyState === WebSocket.OPEN) {
          socket.send(msg);
        }
      };
    }
  }
}
