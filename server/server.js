const Express = require('express');
const Promise = require('bluebird');
const utils = require('./../utils');

const ERROR_MESSAGES = {
	MISSING_PORT: 'Failed to construct server without port',
};

class Server {
	constructor(metadata, opt, logger) {
		logger.info('Starting server component');
		this.port = utils.getPropertyOrError(opt, 'port', ERROR_MESSAGES.MISSING_PORT);
		this.metadata = metadata;
		this.opt = opt;
		this.app = new Express();
		this.logger = logger;
	}

	async init() {
		return Promise.fromCallback(cb => this.app.listen(this.port, cb))
			.then(() => {
				this.logger.info(`Listening on port: ${this.port}`);
				return Promise.resolve();
			}, (err) => {
				const message = `Failed during server initialization with message: ${err.message}`;
				this.logger.error(message);
				throw new Error(message);
			});
	}
}

Server.Errors = ERROR_MESSAGES;

module.exports = Server;
