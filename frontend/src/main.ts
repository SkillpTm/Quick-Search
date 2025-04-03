import "../styles/main.css";
import "../styles/searchicons.css";
import "../styles/component.css";

import { HideWindow, LaunchSearch, LogErrorTS } from "../wailsjs/go/app/App";
import { EventsOn } from "../wailsjs/runtime/runtime";

import { StateHandler } from "./app/statehandler";

const stateHandler = new StateHandler();

// disable right click
document.oncontextmenu = () => {
    return false;
}

// focus the searchBar on load
window.onload = () => {
    stateHandler.uiHandler.searchBar.focus();
    stateHandler.reset();
}

// makes sure the searchBar is always focused
document.addEventListener("click", () => {
    stateHandler.uiHandler.searchBar.focus();
});

// if the input bar is not selected anymore the user selected another window, so we hide
stateHandler.uiHandler.searchBar.addEventListener("blur", () => {
    setTimeout(() => {
        if (document.activeElement === stateHandler.uiHandler.searchBar) {
            HideWindow();
            stateHandler.reset();
        }
      }, 50);
}),

// move the highlighted section with arrow keys and open a file with enter
document.addEventListener("keydown", async (event) => {
    if (event.ctrlKey && event.key === "ArrowUp") {
        event.preventDefault()
        stateHandler.searchMode.updatePage(-1);
        event.preventDefault()
    } else if (event.ctrlKey && event.key === "ArrowDown") {
        stateHandler.searchMode.updatePage(1);
    } else if (event.key === "ArrowUp") {
        event.preventDefault()
        stateHandler.searchMode.updateHighlightedComp(-1);
    } else if (event.key === "ArrowDown") {
        event.preventDefault()
        stateHandler.searchMode.updateHighlightedComp(1);
    } else if (event.key === "Enter") {
        await stateHandler.openFile(stateHandler.searchMode.getHighlightedFile());
    }
});

// send the current input to Go to search the file system
stateHandler.uiHandler.searchBar.addEventListener("input", async () => {
    stateHandler.searchMode.newSearch();
    await LaunchSearch(stateHandler.uiHandler.searchBar.value);
});

// when Go found results receive, handle and display them
EventsOn("searchResult", (results: string[]) => {
    stateHandler.searchMode.newResults(results);
});

// catches all synchronous errors and passes them for error logging to Go
window.onerror = function (_message, source, lineno, colno, error) {
    LogErrorTS(
        `${source ?? "Unknown Source"}: ${lineno ?? "?"}, ${colno ?? "?"}:`.replaceAll("\n", " ") + "\n" +
        `--> ${error?.message ?? "unknown error"}:`.replaceAll("\n", " ") + "\n" +
        `[TS] ${error?.name ?? "Error"}`.replaceAll("\n", " ")
    );
    return true;
};

// catches all Promise errors and passes them for error logging to Go
window.addEventListener("unhandledrejection", (event) => {
    LogErrorTS(
        `${event.reason?.stack ?? "Unknown Source"}`.replaceAll("\n", " ") + "\n" +
        `--> ${event.reason?.message ?? event.reason ?? "unknown error"}`.replaceAll("\n", " ") + "\n" +
        `[TS] Promise Rejection: ${event.reason?.name ?? "Error"}`.replaceAll("\n", " ")
    );
});