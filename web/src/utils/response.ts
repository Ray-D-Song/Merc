// The `responseHandler` is used to handle backend responses
// Now simplified to just return the data directly, as HTTP status codes
// are used to determine success/failure
function responseHandler<T>(response: unknown): T {
  return response as T;
}

export {
  responseHandler
}
