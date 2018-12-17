const rp = require('request-promise');
const _ = require('lodash');
const utils = require('./../utils');

const ERROR_MESSAGES = {
	MISSING_BASE_URL: 'Failed to construct Codefresh API service, missing Codefresh base URL',
	MISSING_TOKEN: 'Failed to construct Codefresh API service, missing Codefresh token',
};

class Codefresh {
	constructor(metadata, options) {
		this.options = options;
		this.metadata = metadata;
		this.defaults = {
			baseUrl: utils.getPropertyOrError(options, 'baseURL', ERROR_MESSAGES.MISSING_BASE_URL),
			headers: {
				Authorization: utils.getPropertyOrError(options, 'token', ERROR_MESSAGES.MISSING_TOKEN),
				'Codefresh-Agent-Version': metadata.version,
			},
			json: true,
			timeout: 30 * 1000,
		};
	}

	async init() {
		return Promise.resolve();
	}

	async _call(options = {}) {
		const opt = _.defaults({}, this.defaults, options);
		return rp(opt);
	}

	fetchTasksToExecute(logger) {
		logger.info('Calling Codefresh API to fetch jobs');
		const url = '/api/agent/tasks';
		return this._call({
			url,
			method: 'GET',
		});
	}

	reportStatus(logger, status) {
		logger.info({ status }, 'Calling Codefresh API to report status');
		const url = '/api/agent/status';
		return this._call({
			url,
			method: 'PUT',
			body: {
				status,
			},
		});
	}
}

Codefresh.Errors = ERROR_MESSAGES;

module.exports = Codefresh;
