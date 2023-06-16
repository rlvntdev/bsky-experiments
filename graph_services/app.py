import asyncio
from os import environ, getenv, getcwd
print(getcwd())
from dotenv import load_dotenv
from flask_socketio import SocketIO
from flask import Flask
from flask_restful import Api


load_dotenv()

def run_socket_io(app, socket_io, host: str, port: int, debug: bool, use_reloader: bool = False, log_output: bool = False):
	socket_io.run(app, host=host, port=port, debug=debug, use_reloader=use_reloader, log_output=log_output)

redis_host = getenv("REDIS_HOST")
redis_port = getenv("REDIS_PORT")
broker_url = f"redis://{redis_host}:{redis_port}"
backend_url = f"redis://{redis_host}:{redis_port}"

import logging
from logging import ERROR, INFO, FileHandler, StreamHandler, Formatter, getLogger
from sys import stdout
def get_logger():
	logger = getLogger()
	if not logger.hasHandlers():
		logger.setLevel(INFO)
		file_handler = FileHandler("graph_services.log")
		file_handler.setLevel(ERROR)
		console_handler = StreamHandler(stdout)
		console_handler.setLevel(INFO)
		formatter = Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
		file_handler.setFormatter(formatter)
		console_handler.setFormatter(formatter)
		logger.addHandler(file_handler)
		logger.addHandler(console_handler)
	return logger

def get_flask_app():
	return Flask("graph_services")

def populate_api(app):
	from api.ping import Ping
	api = Api(app)
	api.add_resource(Ping, "/ping")
	return app

if __name__ == "__main__":
	app = get_flask_app()
	app = populate_api(app)
	logger = get_logger()
	socket_io = SocketIO(app, logger=logger, engineio_logger=True, async_mode="gevent", message_queue=broker_url)
	run_socket_io(app, socket_io, host="localhost", port=5000, debug=True, use_reloader=True, log_output=True)