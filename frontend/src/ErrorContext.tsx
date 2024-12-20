import {createContext, ReactNode, useState} from "react";

const ErrorContext = createContext<{error: any, setError: (error: any) => void}>({error: undefined, setError: () => {}});

export function ErrorContextProvider({ children }: {children: ReactNode}) {
    let [error, setError] = useState<any>(undefined);

    return <ErrorContext.Provider value={{error: error, setError: setError}}>
        {children}
    </ErrorContext.Provider>
}

export default ErrorContext;
