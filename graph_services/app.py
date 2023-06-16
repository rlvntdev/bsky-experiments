import asyncio
from os import environ, getenv

from dotenv import load_dotenv
from flask_socketio import SocketIO

load_dotenv()

def run_socket_io(app, socket_io, host: str, port: int, debug: bool, use_reloader: bool = False, log_output: bool = False):
	socket_io.run(app, host=host, port=port, debug=debug, use_reloader=use_reloader, log_output=log_output)

