const pino = require('pino');

module.exports = {
	create: (metadata, opt = {}) => pino(Object.assign(opt, {
		base: {
			name: metadata.name,
			time: new Date(),
		},
		timestamp: false,
	})),
};
