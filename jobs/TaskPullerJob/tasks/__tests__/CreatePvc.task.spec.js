const _ = require('lodash');
const { create: createLogger } = require('../../../../services/Logger');
const CreatePvcTask = require('../CreatePvc.task');
const { TASK_PRIORITY } = require('../../../../constants');

jest.mock('./../../../../services/Logger');

describe('CreatePvc task unit tests', () => {

	describe('negative', () => {

		it('Should throw an error when call to Kubernetes service is been rejected', () => {
			const logger = createLogger();
			const kubernetesAPIMock = {
				createPvc: jest.fn().mockRejectedValue(new Error('Error!!!')),
			};
			const taskDef = {};
			const task = new CreatePvcTask(_.noop(), kubernetesAPIMock, logger);
			return expect(task.run(taskDef)).rejects.toThrowError('Failed to run task CreatePvc: Error!!!');
		});

		describe('validation', () => {
			it('should throw error in case the task in not valid', () => {
				const logger = createLogger();
				const kubernetesAPIMock = {};
				const taskDef = {};
				const task = new CreatePvcTask(_.noop(), kubernetesAPIMock, logger);
				return expect(task.validate(taskDef)).rejects.toThrowError('child "spec" fails because ["spec" is required]');
			});
		});
	});

	describe('positive', () => {

		it('Should call Kubernetes service', () => {
			const logger = createLogger();
			const spy = jest.fn().mockResolvedValue();
			const kubernetesAPIMock = {
				createPvc: spy,
			};
			const taskDef = {
				spec: {}
			};
			const task = new CreatePvcTask(_.noop(), kubernetesAPIMock, logger);
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
					name: 'pvcName'
				}
			};
			const spy = jest.fn().mockResolvedValue(spyResult);
			const kubernetesAPIMock = {
				createPvc: spy,
			};
			const taskDef = {
				runtime: {},
				dockerDaemon: {},
			};
			const task = new CreatePvcTask(_.noop(), kubernetesAPIMock, logger);
			return expect(task.run(taskDef)).resolves.toEqual(spyResult);
		});

		it('Should have a HIGH priority', () => {
			expect(CreatePvcTask.priority).toBe(TASK_PRIORITY.HIGH);
		});
	});

});
