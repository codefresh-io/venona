const Base = require('../../BaseJob');
const utils = require('../../../utils');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run task TerminateWorkflow, failed to delete pod',
};

class TerminateWorkflowTask extends Base {
	async run(task) {
		this.logger.info('Running TerminateWorkflow task');
		const runtimeSpec = {
			name: utils.getPropertyOrError(task, 'runtime.metadata.name'),
			namespace: utils.getPropertyOrError(task, 'runtime.metadata.namespace'),
		};

		const dindSpec = {
			name: utils.getPropertyOrError(task, 'dockerDaemon.metadata.name'),
			namespace: utils.getPropertyOrError(task, 'dockerDaemon.metadata.namespace'),
		};
		try {
			await this.kubernetesAPI.deletePod(this.logger, runtimeSpec.namespace, runtimeSpec.name);
			await this.kubernetesAPI.deletePod(this.logger, dindSpec.namespace, dindSpec.name);
			return Promise.resolve('OK');
		} catch (err) {
			const message = `${ERROR_MESSAGES.FAILED_TO_EXECUTE_TASK} with message: ${err.message}`;
			this.logger.error(message);
			throw new Error(message);
		}
	}
}
TerminateWorkflowTask.Errors = ERROR_MESSAGES;
module.exports               = TerminateWorkflowTask;
