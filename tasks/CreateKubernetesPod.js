const Base = require('./BaseTask');

const ERROR_MESSAGES = {
	MISSING_RESOURCE_SPEC: 'Failed to run task CreateKubernetesPod, missing pod definitions',
	FAILED_TO_EXECUTE_TASK: 'Failed to run task CreateKubernetesPod, failed to create pod',
};

class CreateKubernetesPod extends Base {
	async run(podDef) {
		this.logger.info('Running CreateKubernetesPod task');
		try {
			const res = await this.kubernetesAPI.createPod(this.logger, podDef);
			return res;
		} catch (err) {
			const message = `${ERROR_MESSAGES.FAILED_TO_EXECUTE_TASK} with message: ${err.message}`;
			this.logger.error(message);
			throw new Error(message);
		}
	}
}
CreateKubernetesPod.Errors = ERROR_MESSAGES;
module.exports = CreateKubernetesPod;
