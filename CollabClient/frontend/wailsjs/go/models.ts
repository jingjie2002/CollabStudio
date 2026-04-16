export namespace main {
	
	export class LANServer {
	    name: string;
	    ip: string;
	    tag: string;
	    recommended: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LANServer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.ip = source["ip"];
	        this.tag = source["tag"];
	        this.recommended = source["recommended"];
	    }
	}

}

