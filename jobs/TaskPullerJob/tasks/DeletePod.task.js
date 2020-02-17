const _ = require('lodash');
const Joi = require('joi');
const Base = require('../../BaseJob');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run task DeletePod',
};

class DeletePodTask extends Base {
	async run(task) {
		this.logger.info('Running DeletePod task');
		try {
			const service = await this.getKubernetesService(_.get(task, 'metadata.reName'));
			await service.deletePod(this.logger, task.spec.namespace, task.spec.name);
		} catch (err) {
			// we treat 404 as if the operation succeeded
			if (_.get(err, 'code') !== 404) {
				const message = `${ERROR_MESSAGES.FAILED_TO_EXECUTE_TASK}: ${err.message}`;
				this.logger.error(message);
				throw new Error(message);
			}
		}

		return Promise.resolve('OK');
	}

	async validate(task) {
		return Joi.validate(task, DeletePodTask.validationSchema);
	}
}

DeletePodTask.Errors = ERROR_MESSAGES;
DeletePodTask.validationSchema = Joi.object().keys({
	spec: Joi.object().keys({
		namespace: Joi.string().required(),
		name: Joi.string().required()
	}).required(),
}).options({ stripUnknown: true });
module.exports       = DeletePodTask;
