(function(){

    function formatError(time) {
        var t = ""
        
        if (time > 60)
            t = Math.round(time/60) + ' min'
        else
            t = time + ' sec'

        // return "Another person recently changed something. To prevent overwriting each other's data please wait until he/she is done. If no additional changes are made you will gain access in " + t + "."
        return "En annan person har nyligen redigerat något på sidan. För att förhindra att ni skriver över varandras ändringar är det bäst att vänta tills hen är klar. Förutsatt att inga fler redigeringar äger rum så kommer du att få tillgång till sidan om " + t + "." 

    }
    
    function displayError(txt) {
        // +1 for good measure
        var time = parseInt(txt) + 1

        sweetAlert("Redigeringskonflikt", formatError(time), "error");
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
    function input(name, placeholder) {
        var node = document.createElement('input')
        
        node.setAttribute('name', name)
        node.setAttribute('placeholder', placeholder)
        node.setAttribute('autocomplete', 'off')
        node.setAttribute('type', 'text')

        return node
    }
    function textarea(name, placeholder) {
        var node = document.createElement('textarea')
        
        node.setAttribute('name', name)
        node.setAttribute('placeholder', placeholder)
        node.setAttribute('autocomplete', 'off')
        node.setAttribute('form', 'programs')

        return node
    }

    function label(name, altText) {
        var node = document.createElement('label')
        node.textContent = name
        node.setAttribute("title", altText);
        return node
    }

    // Generate one program of the form:
    // key { name: "", rss : "", image : "", description : ""}
    function generateProgram(key, program, isExpanded) {
        var div = document.createElement('div'),
            row = document.createElement('div'),
            cont = document.createElement('div'),
            programKey = input('programs[][key]', "<program-id ifylles här>"),
            programName = input('programs[][name]', "<programnamn här>"),
            programRSS = input('programs[][rss]', "<RSS-länk här>"),
            programCategory = input('programs[][category]', "<kategori här>"),
            programInfo = textarea('programs[][description]', "<programbeskrivning här>"),
            programImg = input('programs[][image]', "<länk till programbild här>"),
            btnEdit = document.createElement('button'),
            btnDel = document.createElement('button')

        btnEdit.textContent = 'redigera'
        btnEdit.onclick = function (event) {
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

        btnDel.textContent = 'ta bort'
        btnDel.onclick = function(event) {
            // Because button is inside form we must prevent submit
            event.preventDefault()

            sweetAlert({
                title: "Vill du ta bort programmet?",
                text: "Du behöver skapa '" + key + "' på nytt efter att den tagits bort.",
                type: "warning",
                showCancelButton: true,
                confirmButtonColor: "#DD6B55",
                confirmButtonText: "Ja, ta bort programmet från appen!",
                closeOnConfirm: false,
                html: false
            }, function(){
                swal("Raderad!",
                    "Programmet togs bort utan problem. Klicka på 'spara' för att uppdatera informationen på servern. Uppdatera (F5) denna sida om du tog bort programmet av misstag.",
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
        
        cont.appendChild(label('Programnamn:', "T.ex på 'stan med bettan' eller '7 myror > X elefanter?'"))
        cont.appendChild(programName)

        cont.appendChild(label('Kategori:', "T.ex. samhälle, nöje & kultur, studentliv eller humor"))
        cont.appendChild(programCategory)

        cont.appendChild(label('Programbeskrivning:', "Beskrivning av programmet på svenska eller engelska"))
        cont.appendChild(programInfo)

        cont.appendChild(label('RSS-länk:', "Ser ut som http://www.radioaf.se/program/<PROGRAM-ID>/feed/?post_type=podcasts"))
        cont.appendChild(programRSS)

        cont.appendChild(label('Programbildslänk:', "En enkel länk till programbilden, oftast på formatet http://www.radioaf.se/wp-content/themes/base/library/includes/timthumb....."))
        cont.appendChild(programImg)
        cont.classList.add('program-container')
        if (isExpanded) {
            cont.classList.add('expanded')
            btnEdit.classList.add('expanded')
        }

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

    function newProgram(event, isExpanded){
        // Because button is inside form we must prevent submit
        event.preventDefault()

        var parent = document.getElementById('programs'),
            first = parent.childNodes[2],
            boiler = generateProgram('', {
                name        : '',
                rss         : '',
                image       : '',
                description : '',
                category    : ''
            }, isExpanded)
        parent.insertBefore(boiler, first)
    }

    window.onload = function () {
        fetchPrograms(generatePrograms)

        var res = document.querySelector('iframe')
        res.onload = function(){
            flashPostResults(res)
        }

        document.getElementById('btn-add').onclick = function(event) {
            newProgram(event, true);
        }
    }

})();
