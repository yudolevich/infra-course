from http.server import HTTPServer, BaseHTTPRequestHandler
from os import getenv

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        file = getenv("FILE", "none")
        self.wfile.write(f"file: {file}\n".encode())
        if file != "none":
            self.wfile.write(open(file, "rb").read())

HTTPServer(('', 8888), Handler).serve_forever()
