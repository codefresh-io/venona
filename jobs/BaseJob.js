class Job {
	constructor(codefreshAPI, kubernetesAPI, logger) {
		this.codefreshAPI = codefreshAPI;
		this.kubernetesAPI = kubernetesAPI;
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
}

module.exports = Job;
