import {useState} from 'react';
import './App.css';
import {RunCommonFixes, RunDebug} from "../wailsjs/go/backend/App";
import {Card, CardDescription, CardFooter, CardHeader, CardTitle} from "@/components/ui/card";
import {Button} from "@/components/ui/button";
import {LoadingSpinner} from "@/components/ui/spinner";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger
} from "@/components/ui/alert-dialog";
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle
} from "@/components/ui/dialog";
import {Label} from "@/components/ui/label";
import {Input} from "@/components/ui/input";
import {CopyIcon, ExclamationTriangleIcon} from "@radix-ui/react-icons";
import {BrowserOpenURL} from "../wailsjs/runtime";
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert";

function App() {
    const [debugCode, setDebugCode] = useState('');
    const [debugRunning, setDebugRunning] = useState(false);
    const [showDebugCode, setShowDebugCode] = useState(false);

    const [fixesRunning, setFixesRunning] = useState(false);

    const [error, setError] = useState('');

    function runDebug() {
        setDebugRunning(true);

        RunDebug().then((result) => {
            setDebugRunning(false);
            setDebugCode(`dbg:${result}`);
            setShowDebugCode(true);
        }).catch((err) => {
            setDebugRunning(false);
            console.error(err);
            setError(err);
        })
    }

    function runFixes() {
        setFixesRunning(true);

        RunCommonFixes().then((result) => {
            setFixesRunning(false);
            console.log(result);
        }).catch((err) => {
            setFixesRunning(false);
            console.error(err);
            setError(err);
        })
    }

    return (
        <div className="flex flex-col gap-4">
            <div className="flex gap-4">
                <Card className="w-[300px] flex flex-col">
                    <CardHeader>
                        <CardTitle>Diagnostics</CardTitle>
                        <CardDescription>Run a diagnostics check and generate a debug code</CardDescription>
                    </CardHeader>
                    <CardFooter className="mt-auto">
                        <Button onClick={runDebug} disabled={debugRunning || fixesRunning}>Run</Button>
                    </CardFooter>
                </Card>
                <Card className="w-[300px] flex flex-col">
                    <CardHeader>
                        <CardTitle>Fix Common Issues</CardTitle>
                        <CardDescription>Fixes common issues with the app</CardDescription>
                    </CardHeader>
                    <CardFooter className="mt-auto">
                        <AlertDialog>
                            <AlertDialogTrigger asChild>
                                <Button variant="destructive" disabled={debugRunning || fixesRunning}>Run</Button>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                                <AlertDialogHeader>
                                    <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                                    <AlertDialogDescription>
                                        This actions may break any existing instances you have installed via the app, if
                                        this happens you will need to repair those instances by following the guide
                                        below.
                                        <br/><br/><u><span className="cursor-pointer"
                                                           onClick={() => BrowserOpenURL("https://docs.feed-the-beast.com/docs/app/Instances/repair")}>https://docs.feed-the-beast.com/docs/app/Instances/repair</span></u>
                                    </AlertDialogDescription>
                                </AlertDialogHeader>
                                <AlertDialogFooter>
                                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                                    <AlertDialogAction onClick={runFixes}>Continue</AlertDialogAction>
                                </AlertDialogFooter>
                            </AlertDialogContent>
                        </AlertDialog>
                    </CardFooter>
                </Card>
            </div>

            { error !== "" &&
                <Alert className="flex-1" variant="destructive">
                    <ExclamationTriangleIcon className="h-4 w-4"/>
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>
                        {error}
                    </AlertDescription>
                </Alert>
            }
            <Dialog open={debugRunning || fixesRunning}>
                <DialogContent className="sm:max-w-md [&>button]:hidden">
                    <DialogHeader>
                        <DialogTitle>Please wait...</DialogTitle>
                        <DialogDescription>
                            Please wait while the debug tool runs its checks.
                        </DialogDescription>
                        <div className="flex flex-col items-center justify-center">
                            <LoadingSpinner/>
                        </div>
                    </DialogHeader>
                </DialogContent>
            </Dialog>

            <Dialog open={showDebugCode} onOpenChange={setShowDebugCode}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>Share link</DialogTitle>
                        <DialogDescription>
                            Please share this link with the support team to help diagnose your issue.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex items-center space-x-2">
                        <div className="grid flex-1 gap-2">
                            <Label htmlFor="link" className="sr-only">
                                Debug Code
                            </Label>
                            <Input
                                id="debugCode"
                                defaultValue={debugCode}
                                readOnly
                            />
                        </div>
                        <Button type="submit" size="sm" className="px-3">
                            <span className="sr-only">Copy</span>
                            <CopyIcon className="h-4 w-4"/>
                        </Button>
                    </div>
                    <DialogFooter className="sm:justify-start">
                        <DialogClose asChild>
                            <Button type="button" variant="secondary">
                                Close
                            </Button>
                        </DialogClose>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    )
}

export default App
