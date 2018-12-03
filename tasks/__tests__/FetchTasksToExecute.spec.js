const _ = require('lodash');
const { create: createLogger } = require('../../services/Logger');
const FetchTasksToExecute = require('../FetchTasksToExecute');

jest.mock('./../../services/Logger');

describe('FetchTasksToExecute unit tests', () => {
	it('Should throw an error when codefresh service call failed', () => {
		const logger = createLogger();
		const task = new FetchTasksToExecute({
			fetchTasksToExecute: jest.fn().mockRejectedValue(new Error('Failed')),
		}, _.noop(), logger);
		return expect(task.run()).rejects.toThrowError('Failed to run task FetchTasksToExecute, call to Codefresh rejected with message');
	});

	it('Should log an error when codefresh service call failed', () => {
		const logger = createLogger();
		const task = new FetchTasksToExecute({
			fetchTasksToExecute: jest.fn().mockRejectedValue(new Error('Failed')),
		}, _.noop(), logger);
		return task.run()
			.catch(() => {
				expect(logger.child.mock.results[1].value.error.mock.calls[0][0]).toMatch('Failed to run task FetchTasksToExecute, call to Codefresh rejected with message');
			});
	});

	it('Should pass logger to codefresh api service', () => {
		const spy = jest.fn().mockResolvedValue();
		const logger = createLogger();
		const task = new FetchTasksToExecute({
			fetchTasksToExecute: spy,
		}, _.noop(), logger);
		return task.run()
			.then(() => {
				expect(spy).toHaveBeenCalledWith(expect.objectContaining({
					error: expect.any(Function),
					info: expect.any(Function),
					child: expect.any(Function),
				}));
			});
	});
});
