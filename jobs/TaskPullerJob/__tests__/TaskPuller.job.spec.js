const _ = require('lodash');
const { create: createLogger } = require('../../../services/Logger');
const TaskPullerJob = require('../TaskPuller.job');
const CreatePodTask = require('../tasks/CreatePod.task');

jest.mock('./../../../services/Logger');
jest.mock('./../tasks/CreatePod.task');

describe('TaskPullerJob unit tests', () => {
	it('Should throw an error when codefresh service call failed', () => {
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: jest.fn().mockRejectedValue(new Error('Failed')),
		}, _.noop(), logger);
		return expect(task.exec()).rejects.toThrowError('Failed to run job TaskPuller, call to Codefresh rejected with message');
	});

	it('Should pass logger to codefresh api service', () => {
		const spy = jest.fn().mockResolvedValue();
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: spy,
		}, _.noop(), logger);
		return task.exec()
			.then(() => {
				expect(spy).toHaveBeenCalledWith(expect.objectContaining({
					error: expect.any(Function),
					info: expect.any(Function),
					child: expect.any(Function),
				}));
			});
	});

	it('Should map all results to tasks by types and execute them', () => {
		CreatePodTask.mockImplementationOnce(() => {
			return {
				exec: jest.fn(() => {
					return {
						status: 'ok'
					};
				}),
			};
		});
		const tasks = [
			{
				type: 'CreatePod',
			}
		];
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: jest.fn().mockResolvedValue(tasks),
		}, _.noop(), logger);
		return expect(task.exec()).resolves.toEqual([{ status: 'ok'}]);
	});

	it('Should not fail when unknown type task arrives', () => {
		const tasks = [
			{
				type: 'Fake-Task-Type',
			},
		];
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: jest.fn().mockResolvedValue(tasks),
		}, _.noop(), logger);
		return expect(task.exec()).resolves.toEqual([]);
	});

	it('Should not fail when non-typed task arrives', () => {
		const tasks = [
			{
				prop: 'a'
			},
		];
		const logger = createLogger();
		const task = new TaskPullerJob({
			pullTasks: jest.fn().mockResolvedValue(tasks),
		}, _.noop(), logger);
		return expect(task.exec()).resolves.toEqual([]);
	});
});
