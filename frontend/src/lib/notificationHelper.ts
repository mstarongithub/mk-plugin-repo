import toast, { type Renderable, type ValueOrFunction } from "svelte-french-toast";

const className = '!btn';

export const notify = {
    success: (message : string) => {
        return toast.success(message, {
            className
        });
    },
    error: (message : string) => {
        return toast.error(message, {
            className
        });
    },
    loading: (message : string) => {
        return toast.loading(message, {
            className
        });
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    promise: (promise: Promise<unknown>, msgs: { loading: Renderable, success: ValueOrFunction<Renderable, unknown>, error: ValueOrFunction<Renderable, any> }) => {
        return toast.promise(promise, msgs, {
            className
        });
    }
}