export abstract class AppError extends Error {

  constructor(message: string) {
    super(message);
    this.name = this.constructor.name;
    Object.setPrototypeOf(this, new.target.prototype); // keep instanceof
  }
}

export const ErrorName = {
  Unknown: 'Unknown',
  InternalServerError: 'InternalServerError',
  InvalidParameter: 'InvalidParameter',
  MissingParameter: 'MissingParameter',

  NetworkError: 'NetworkError',
  RequestTimeout: 'RequestTimeout',
  ConnectionFailed: 'ConnectionFailed',

  Unauthorized: 'Unauthorized',
  Forbidden: 'Forbidden',
  TokenExpired: 'TokenExpired',
  InvalidCredentials: 'InvalidCredentials',

  ResourceNotFound: 'ResourceNotFound',
  ResourceAlreadyExists: 'ResourceAlreadyExists',
  OperationNotAllowed: 'OperationNotAllowed',
  ValidationFailed: 'ValidationFailed',

  DataParseError: 'DataParseError',
  StorageError: 'StorageError',
} as const;

export type ErrorName = typeof ErrorName[keyof typeof ErrorName];

type ErrorCtor = new (message?: string) => AppError;

export const Errors = {} as Record<ErrorName, ErrorCtor>;

Object.values(ErrorName).forEach(name => {
  const Ctor = {
    [name]: class extends AppError {
      constructor(message?: string) {
        super(message ?? name);
      }
    },
  }[name];

  Errors[name as ErrorName] = Ctor;
});
