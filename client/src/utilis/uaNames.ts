const arr = ["mimi", "babo"]




function getAgentInfo(userAgent: string){
    const browser = getBrowser(userAgent);
    const os = getOS(userAgent);
    if(browser && os){
        return `${os} | ${browser}`;
    }else if(os){
        return os;
    }else{
        return browser;
    }
}

function getBrowser(userAgent: string){
    switch(true){
        case userAgent.includes("Firefox"):
            return "Firefox";
        case userAgent.includes("Chrome"):
            return "Chrome";
        case userAgent.includes("Safari"):
            return "Safari";
        case userAgent.includes("Opera") || userAgent.includes("OPR"):
            return "Opera";
        case userAgent.includes("Edg"):
            return "Edge";
        default:
            return null;
    }
}

function getOS(userAgent: string){
    switch(true){
        case userAgent.includes("Windows"):
            return "Windows";
        case userAgent.includes("Android"):
            return "Android";
        case userAgent.includes("Macintosh"):
            return "macOS";
        case userAgent.includes("Iphone") || userAgent.includes("iPad"):
            return "iOS";
        case userAgent.includes("X11"):
            return "Linux";
        default:
            return null;
    }
    
}


