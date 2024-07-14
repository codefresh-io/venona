const _ = require('lodash');
const { create: createLogger } = require('../../../../services/Logger');
const CreatePodTask = require('../CreatePod.task');
const { TASK_PRIORITY } = require('../../../../constants');

jest.mock('./../../../../services/Logger');

describe('CreatePod task unit tests', () => {

	describe('negative', () => {

		it('Should throw an error when call to Kubernetes service is been rejected', () => {
			const logger = createLogger();
			const kubernetesAPIMock = {
				createPod: jest.fn().mockRejectedValue(new Error('Error!!!')),
			};
			const taskDef = {};
			const task = new CreatePodTask(_.noop(), kubernetesAPIMock, logger);
			return expect(task.run(taskDef)).rejects.toThrowError('Failed to run task CreatePod: Error!!!');
		});
	});

	describe('positive', () => {

		it('Should call twice to Kubernetes service', () => {
			const logger = createLogger();
			const spy = jest.fn().mockResolvedValue();
			const kubernetesAPIMock = {
				createPod: spy,
			};
			const taskDef = {
				spec: {}
			};
			const task = new CreatePodTask(_.noop(), kubernetesAPIMock, logger);
			return task.run(taskDef)
				.then(() => {
					const loggerMacher = expect.objectContaining({
						error: expect.any(Function),
						info: expect.any(Function),
						child: expect.any(Function),
					});
					expect(spy).toHaveBeenCalledTimes(1);
					expect(spy).toHaveBeenNthCalledWith(1, loggerMacher, taskDef.spec);
				});
		});

		it('Should return value', () => {
			const logger = createLogger();
			const spyResult = {
				metadata: {
					name: 'podName'
				}
			};
			const spy = jest.fn().mockResolvedValue(spyResult);
			const kubernetesAPIMock = {
				createPod: spy,
			};
			const taskDef = {
				runtime: {},
				dockerDaemon: {},
			};
			const task = new CreatePodTask(_.noop(), kubernetesAPIMock, logger);
			return expect(task.run(taskDef)).resolves.toEqual(spyResult);
		});

		it('Should have a HIGH priority', () => {
			expect(CreatePodTask.priority).toBe(TASK_PRIORITY.HIGH);
		});
	});

});
