const _ = require('lodash');

class Job {
	constructor(codefreshAPI, runtimes, logger) {
		this.codefreshAPI = codefreshAPI;
		this.runtimes = runtimes;
		this.logger = logger;
	}

	async exec(task) {
		await this.validate(task);
		return this.run(task);
	}

	async run() {
		throw new Error('not implemented');
	}

	async validate() {
		throw new Error('not implemented');
	}

	async getKubernetesService(runtime) {
		if (_.has(this, `runtimes[${runtime}]`)) {
			return this.runtimes[runtime].kubernetesAPI;
		}
		throw new Error(`Kubernetes client for runtime ${runtime} was not found`);
	}
}

module.exports = Job;
