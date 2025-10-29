from http.server import BaseHTTPRequestHandler, HTTPServer
import logging

class SimpleHandler(BaseHTTPRequestHandler):
    def log_request_info(self):
        content_length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_length).decode('utf-8') if content_length else ''
        logging.info(f"Path: {self.path}")
        logging.info(f"Method: {self.command}")
        logging.info(f"Headers: {self.headers}")
        logging.info(f"Body: {body}")

    def do_GET(self):
        self.log_request_info()
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"OK")

    def do_POST(self):
        self.log_request_info()
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"OK")

logging.basicConfig(level=logging.INFO)
server = HTTPServer(("0.0.0.0", 1080), SimpleHandler)
print("Server running on http://localhost:1080")
server.serve_forever()
