
class GoInvictaDuxObject {

    MarshalArray(obj) {
        var arr = [];

        switch (typeof obj[0]) {
            case 'number':
                return `[${obj.join(',')}]`;
            case 'string':
                for (var i = 0; i < obj.length; i++) {
                    arr.push(`"${obj[i]}"`);
                }
                break;
            case 'boolean':
                for (var i = 0; i < obj.length; i++) {
                    arr.push((obj[i]) ? '+': '');
                }
                break;
            case 'object':
                if (obj[0] instanceof Array) {
                    for (var i = 0; i < obj.length; i++) {
                        arr.push(this.MarshalArray(obj[i]));
                    }
                } else {
                    for (var i = 0; i < obj.length; i++) {
                        arr.push(this.Marshal(obj[i]));
                    }
                }
                break;
        }
    
        return `[${arr.join(',')}]`;
    }
    
    Marshal(obj) {
        var arr = [];
    
        for (const key in obj) {
            if (obj[key] === null || obj[key] === undefined) {
                console.log('null');
            }
    
            switch (typeof obj[key]) {
                case 'number':
                    arr.push(obj[key]);
                    break;
                case 'string':
                    arr.push(`"${obj[key]}"`);
                    break;
                case 'boolean':
                    arr.push((obj[key]) ? '+': '');
                    break;
                case 'object':
                    if (obj[key] instanceof Array) {
                        arr.push(this.MarshalArray(obj[key]));
                    } else {
                        arr.push(this.Marshal(obj[key]));
                    }
                    break;
                default:
                    console.log("unexpected type");
            }
        }
    
        return `{${arr.join(',')}}`;
    }

    ValidateStringArray(data) {
        var i = 0;
        var start_i = 0;
        var inString = false;
        var arr = [];
    
        while (i < data.length) {
            switch (data[i]) {
                case ',':
                case ']':
                    if (!inString) {
                        arr.push(data.substring(start_i+1,i-1));
                    }
                    break;
                case '"':
                    if (!inString) {
                        inString = true;
                        start_i = i;
                    } else {
                        inString = false;
                    }
                    break;
                case '\\':
                    i++;
                    break;
            }
            i++;
        }
    
        return arr;
    }
    
    ValidateNestedArray(data, objArr) {
        var i = 0;
        var start_i = 0;
        var inArray = 0;
        var inString = false;
        var arr = [];

        while (i < data.length) {
            switch(data[i]) {
                case ',':
                    if (inArray == 0 && !inString) {
                        var ev = data.substring(start_i, i);
                        var result = this.ValidateArray(ev,objArr[0]);
                        arr.push(result);
                    }
                    break;
                case '"':
                    if (!inString) {
                        inString = true;
                    } else {
                        inString = false;
                    }
                    break;
                case '\\':
                    i++;
                    break;
                case '[':
                    if (!inString) {
                        if (inArray == 0) {
                            start_i = i;
                        }

                        inArray++;
                    }
                    break;
                case ']':
                    if (!inString) {
                        inArray--;
                    }
                    break;
            }
            i++;
        }

        if (i+1 >= data.length) {
            var ev = data.substring(start_i, data.length);
            var result = this.ValidateArray(ev,objArr[0]);
            arr.push(result);
        }

        return arr;
    }

    ValidateObjectArray(data, objArr) {
        var i = 0;
        var start_i = 0;
        var inObject = 0;
        var inString = false;
        var arr = [];

        while (i < data.length) {
            switch(data[i]) {
                case ',':
                case ']':
                    if (inObject == 0 && !inString) {
                        var ev = data.substring(start_i,i);
                        var obj = {...objArr};
                        this.Unmarshal(ev, obj);
                        arr.push(obj);
                    }
                    break;
                case '"':
                    if (!inString) {
                        inString = true;
                    } else {
                        inString = false;
                    }
                    break;
                case '\\':
                    i++;
                    break;
                case '{':
                    if (!inString) {
                        if (inObject == 0) {
                            start_i = i;
                        }
                        inObject++;
                    }
                    break;
                case '}':
                    if (!inString) {
                        inObject--;
                    }
                    break;
            }
            i++;
        }

        return arr;
    }
    
    ValidateArray(data, objArr) {
        switch (typeof objArr) {
            case 'number':
                if (Number.isInteger(objArr)) {
                    var arr = [];
                    for (const v of data.substring(1, data.length-1).split(',')) {
                        arr.push(parseInt(v));
                    }
                    return arr;
                } else {
                    var arr = [];
                    for (const v of data.substring(1, data.length-1).split(',')) {
                        arr.push(parseFloat(v));
                    }
                    return arr;
                }
            case 'string':
                return this.ValidateStringArray(data);
            case 'boolean':
                var arr = [];
                for (const v of data.substring(1, data.length-1).split(',')) {
                    if (v === '+') {
                        arr.push(true);
                    } else if (v.length == 0) {
                        arr.push(false);
                    }
                }
                return arr;
            case 'object':
                if (objArr instanceof Array) {
                    return this.ValidateNestedArray(data.substring(1,data.length-1), objArr);
                } else {
                    return this.ValidateObjectArray(data, objArr);
                }
        }
    }
    
    Validate(data, obj, key) {
        switch (typeof obj[key]) {
            case 'number':
                if (Number.isInteger(obj[key])) {
                    obj[key] = parseInt(data);
                } else {
                    obj[key] = parseFloat(data);
                }
                break;
            case 'string':
                obj[key] = data.substring(1, data.length-1);
                break;
            case 'boolean':
                if (data === '+') {
                    obj[key] = true;
                } else if (data.length == 0) {
                    obj[key] = false;
                }
            case 'object':
                if (obj[key] instanceof Array) {
                    obj[key] = this.ValidateArray(data, obj[key][0]);
                } else {
                    this.Unmarshal(data, obj[key]);
                }
                break;
        }
    }
    
    Unmarshal(data, obj) {
        const keysArray = Object.keys(obj);
    
        var i = 1;
        var elm = 0;
        var start_i = 1;
        var inString = false;
        var isEscaped = false;
        var inArray = 0;
        var inObject = 0;
    
        while (i < data.length) {
            switch (data[i]) {
                case ',':
                    if (inArray == 0 && inObject == 0 && !inString) {
                        var ev = data.substring(start_i,i);
                        this.Validate(ev, obj, keysArray[elm]);
    
                        start_i = i + 1
                        elm++;
                    }
                    break;
                case '"':
                    if (inObject == 0 && inArray == 0) {
                        if (isEscaped) {
                            isEscaped = false;
                            i++;
                            continue;
                        }
    
                        if (!inString) {
                            start_i = i;
                            inString = true;
                        } else {
                            inString = false;
                        }
                    }
                    break;
                case '\\':
                    isEscaped = true;
                    break;
                case '{':
                    if (!inString && inArray == 0) {
                        if (inObject == 0) {
                            start_i = i
                        }
                        inObject++;
                    }
                    break;
                case '}':
                    if (inObject > 0) {
                        inObject--;
                    }
                    break;
                case '[':
                    if (inObject == 0 && !inString) {
                        inArray++;
                    }
                    break;
                case ']':
                    if (inObject == 0 && !inString) {
                        inArray--;
                    }
                    break;
            }
    
            if (i+1 >= data.length) {
                var ev = data.substring(start_i, data.length-1);
                this.Validate(ev, obj, keysArray[elm]);
            }
            i++;
        }
    }
}

window.GIDO = new GoInvictaDuxObject();