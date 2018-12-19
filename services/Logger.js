const pino = require('pino');

module.exports = {
	create: (metadata, opt = {}) => pino(Object.assign(opt, {
		base: {
			time: new Date(),
		},
		timestamp: false,
	})),
};
