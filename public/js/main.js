(function(){

    function formatError(time) {
        var t = ""
        
        if (time > 60)
            t = Math.round(time/60) + ' min'
        else
            t = time + ' sec'

        return "Another person recently changed something. To prevent overwriting each other's data please wait until he/she is done. If no additional changes are made you will gain access in " + t + "."

    }
    
    function displayError(txt) {
        // +1 for good measure
        var time = parseInt(txt) + 1

        sweetAlert("Edit Conflict", formatError(time), "error");
        setTimeout(function countdown() {
            var el = document.querySelector(".sweet-alert > p")
            el.textContent = formatError(--time)
            if (time > 0)
                setTimeout(countdown, 1000)
            else
                window.location.reload(true)        
        }, 1000)
    }

    // Classic XHR to get programs.json
    function fetchPrograms (callb) {
        var xhr = new XMLHttpRequest()

        xhr.open('GET', '/programs')

        xhr.onreadystatechange = function() {
            if (xhr.readyState === 4) {
                if (xhr.status !== 200)
                    return displayError(xhr.responseText)

                callb(JSON.parse(xhr.responseText))
            }
        }
        xhr.send()
    }

    // Shorthand to make the input elements
    function input(name) {
        var node = document.createElement('input')
        
        node.setAttribute('name', name)
        node.setAttribute('autocomplete', 'off')
        node.setAttribute('type', 'text')

        return node
    }
    function textarea(name) {
        var node = document.createElement('textarea')
        
        node.setAttribute('name', name)
        node.setAttribute('autocomplete', 'off')
        node.setAttribute('form', 'programs')

        return node
    }

    function label(name){
        var node = document.createElement('label')
        node.textContent = name
        return node
    }

    // Generate one program of the form:
    // key { name: "", rss : "", image : "", description : ""}
    function generateProgram(key, program) {
        var div = document.createElement('div'),
            row = document.createElement('div'),
            cont = document.createElement('div'),
            programKey = input('programs[][key]'),
            programName = input('programs[][name]'),
            programRSS = input('programs[][rss]'),
            programCategory = input('programs[][category]'),
            programInfo = textarea('programs[][description]'),
            programImg = input('programs[][image]'),
            btnEdit = document.createElement('button'),
            btnDel = document.createElement('button')

        btnEdit.textContent = 'edit'
        btnEdit.onclick = function () {
            // Because button is inside form we must prevent submit
            event.preventDefault()

            // Animate using max height. Should be set to something
            // larger than container will ever be, but not too large
            // because the animation will animate over all pixels
            if (cont.classList.contains('expanded')) {
                cont.classList.remove('expanded')
                btnEdit.classList.remove('expanded')
            } else {
                cont.classList.add('expanded')
                btnEdit.classList.add('expanded')
            }
        }

        btnDel.textContent = 'delete'
        btnDel.onclick = function() {
            // Because button is inside form we must prevent submit
            event.preventDefault()

            sweetAlert({
                title: "Are you sure?",
                text: "You need to manually recover '" + key + "' once deleted.",
                type: "warning",
                showCancelButton: true,
                confirmButtonColor: "#DD6B55",
                confirmButtonText: "Yes, delete it!",
                closeOnConfirm: false,
                html: false
            }, function(){
                swal("Deleted!",
                    "Program was deleted for you. Click 'save' to update server as well. Refresh this web page if you misclicked.",
                    "success");
                div.parentNode.removeChild(div)
            });
        }
        
        programKey.value = key
        programName.value = program.name
        programRSS.value = program.rss
        programImg.value = program.image
        programCategory.value = program.category
        programInfo.value = program.description
        
        cont.appendChild(label('Name:'))
        cont.appendChild(programName)

        cont.appendChild(label('Category:'))
        cont.appendChild(programCategory)

        cont.appendChild(label('Description:'))
        cont.appendChild(programInfo)

        cont.appendChild(label('RSS:'))
        cont.appendChild(programRSS)

        cont.appendChild(label('Image:'))
        cont.appendChild(programImg)
        cont.classList.add('program-container')

        row.appendChild(programKey)
        row.appendChild(btnEdit)
        row.appendChild(btnDel)
        row.classList.add('program-header')

        div.appendChild(row)
        div.appendChild(cont)
        div.classList.add('program')

        return div
    }

    function generatePrograms(programs){
        var keys = Object.keys(programs),
            parent = document.getElementById('programs')

        keys
            .map(function (k) {
                return generateProgram(k, programs[k])
            })
            .forEach(function (div) {
                parent.appendChild(div)
            })
    }

    function flashPostResults(frame){
        var flashes = 3

        var white = function(){
            frame.style.background = '#AAFFAA'
            if (flashes-- > 0)
                setTimeout(blue, 100)
        }
        
        var blue = function(){
            frame.style.background = '#E0EBF5'
            if (flashes-- > 0)
                setTimeout(white, 100)
        }

        white()
    }

    function newProgram(event){
        // Because button is inside form we must prevent submit
        event.preventDefault()

        var parent = document.getElementById('programs'),
            first = parent.childNodes[2],
            boiler = generateProgram('FIXME: program key', {
                name        : 'FIXME: program name',
                rss         : 'FIXME: program rss link',
                image       : 'FIXME: program image link',
                description : 'FIXME: program description',
                category    : 'FIXME: program category'
            })
        parent.insertBefore(boiler, first)
    }

    window.onload = function () {
        fetchPrograms(generatePrograms)

        var res = document.querySelector('iframe')
        res.onload = function(){
            flashPostResults(res)
        }

        document.getElementById('btn-add').onclick = newProgram
    }

})();