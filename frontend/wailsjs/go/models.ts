export namespace main {
	
	export class AllVersions {
	    zdp: string;
	    zcc: string;
	    zep: string;
	
	    static createFrom(source: any = {}) {
	        return new AllVersions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.zdp = source["zdp"];
	        this.zcc = source["zcc"];
	        this.zep = source["zep"];
	    }
	}
	export class ClassifierOutput {
	    command: string;
	    output: string;
	    ocrTextPath: string;
	    extractedTextPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ClassifierOutput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.command = source["command"];
	        this.output = source["output"];
	        this.ocrTextPath = source["ocrTextPath"];
	        this.extractedTextPath = source["extractedTextPath"];
	    }
	}

}

