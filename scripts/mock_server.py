from http.server import BaseHTTPRequestHandler, HTTPServer
import sys

class RequestHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        print(f"\n[Mock Server] Received POST request to {self.path}")
        print(f"Headers: {self.headers}")
        print(f"Body:\n{post_data.decode('utf-8')}\n")
        
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"OK")

    def log_message(self, format, *args):
        return # Squelch standard logging to keep output clean

def run(server_class=HTTPServer, handler_class=RequestHandler, port=9000):
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    print(f"Starting mock HTTP server on port {port}...")
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        pass
    httpd.server_close()
    print("Stopping httpd...")

if __name__ == '__main__':
    run()
