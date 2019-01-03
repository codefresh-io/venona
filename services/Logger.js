const pino = require('pino');

class Logger {

}

module.exports = Logger;

module.exports = {
	create: (metadata, opt = {}) => pino(Object.assign(opt, {
		base: {
			time: new Date(),
		},
		timestamp: false,
	})),
};
