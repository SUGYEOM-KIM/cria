export namespace vcs {
	
	export class UpgradeHistory {
	    version: string;
	    hash: string;
	    message: string;
	    date: string;
	    time: string;
	    isAutoUpgrade: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpgradeHistory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.hash = source["hash"];
	        this.message = source["message"];
	        this.date = source["date"];
	        this.time = source["time"];
	        this.isAutoUpgrade = source["isAutoUpgrade"];
	    }
	}

}

