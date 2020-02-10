const Joi = require('joi');
const Base = require('../../BaseJob');
const _ = require('lodash');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run task CreatePod',
};

class CreatePodTask extends Base {
	async run(task) {
		this.logger.info('Running CreatePod task');
		try {
			const pod = await this.kubernetesAPI.createPod(this.logger, task.spec, _.get(task, 'metadata.reName'));
			return pod;
		} catch (err) {
			const message = `${ERROR_MESSAGES.FAILED_TO_EXECUTE_TASK}: ${err.message}`;
			this.logger.error(message);
			throw new Error(message);
		}
	}

	async validate(task) {
		return Joi.validate(task, CreatePodTask.validationSchema);
	}
}
CreatePodTask.Errors = ERROR_MESSAGES;
CreatePodTask.validationSchema = Joi.object().keys({
	spec: Joi.object().required(),
}).options({ stripUnknown: true });
module.exports       = CreatePodTask;
