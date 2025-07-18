const stdin = 0;

function streamComponent() {
  return {
    connected: false,
    status: "Disconnected",
    data: "",
    socket: null,

    connect(url) {
      this.socket = new WebSocket(url);

      this.socket.onopen = () => {
        this.connected = true;
        this.status = "Connected";
      };

      this.socket.onmessage = (e) => {
        this.data = this.data + atob(JSON.parse(e.data).Data)
      };

      this.socket.onclose = () => {
        this.connected = false;
        this.status = "Disconnected";
      };

      this.socket.onerror = () => {
        this.connected = false;
        this.status = "Error";
        this.sendMessage = null;
      };

      this.sendMessage = (msg) => {
        if (this.socket.readyState === WebSocket.OPEN) {
          this.socket.send(JSON.stringify({
            stream: stdin,
            data: btoa(msg + "\n"),
          }));
        }
      };

      this.close = (msg) => {
        this.socket.close();
      };
    }
  }
}
