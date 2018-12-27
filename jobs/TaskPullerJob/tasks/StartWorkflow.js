const Base = require('../../BaseJob');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run task StartWorkflow, failed to create pod',
};

class StartWorkflow extends Base {
	async run(task) {
		this.logger.info('Running StartWorkflow task');
		try {
			const runtime = await this.kubernetesAPI.createPod(this.logger, task.runtime);
			const dind = await this.kubernetesAPI.createPod(this.logger, task.dockerDaemon);
			return {
				runtime,
				dind,
			};
		} catch (err) {
			const message = `${ERROR_MESSAGES.FAILED_TO_EXECUTE_TASK} with message: ${err.message}`;
			this.logger.error(message);
			throw new Error(message);
		}
	}
}
StartWorkflow.Errors = ERROR_MESSAGES;
module.exports = StartWorkflow;
