const Chance = require('chance');

class Job {
	constructor(codefreshAPI, kubernetesAPI, logger) {
		this.codefreshAPI = codefreshAPI;
		this.kubernetesAPI = kubernetesAPI;
		this.logger = logger.child({
			uid: new Chance().guid(),
		});
	}
}

module.exports = Job;
