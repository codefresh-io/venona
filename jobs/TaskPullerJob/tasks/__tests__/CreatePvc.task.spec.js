const _ = require('lodash');
const { create: createLogger } = require('../../../../services/Logger');
const CreatePvcTask = require('../CreatePvc.task');

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
	});

});
