const pino = require('pino');

class Logger {

}

module.exports = Logger;

module.exports = {
	create(metadata, opt = {}) {
		const logger = pino(Object.assign(opt, {
			base: {
				time: new Date(),
			},
			timestamp: false,
		}));

		logger.infoVerbose = (msg) => {
			if (process.env.verbose) {
				logger.info(msg);
			}
		};

		const Child = logger.child.bind(logger);
		logger.child = (opts) => {
			const childLogger = Child(opts);

			childLogger.infoVerbose = (msg) => {
				if (process.env.verbose) {
					childLogger.info(msg);
				}
			};

			return childLogger;
		};

		return logger;
	},
};
