declare var pdfjsLib: any;
const pdfCanvasID = "pdf-canvas",
    fullscreenID = "toggle-fullscreen",
    prevPageID = "prev-page",
    nextPageID = "next-page",
    pageInputID = "page-input",
    pageSpanID = "page-span";

const pdfCanvasHTML = `
<div id="canvas-div">
    <button id="${fullscreenID}">Toggle Full Screen</button>
    <button id="${prevPageID}">Previous</button>
    <input type="number" id="${pageInputID}">
    <span id="${pageSpanID}">/ ?</span>
    <button id="${nextPageID}">Next</button>
</div>

<canvas id="${pdfCanvasID}"></canvas>`

const nonTextFileTypes = [
    // Image types
    "png", "jpg", "jpeg", "svg",
    // Audio types
    "wav", "mp3",
    // Video types
    "mp4", "webm",
    // Other
    "pdf",
]

let pdfDoc = null,
    pdfCtx = null,
    pdfFullScreen: boolean = false,
    pdfCanvas: HTMLCanvasElement = null,
    pageNum: number = 1,
    pageRendering: boolean = false,
    pageNumPending: number = null;

const getExt = (filename: string) => {
    let extSplit = filename.split(".");
    return extSplit[extSplit.length - 1].toLowerCase();
}

const getMimetypeFromExt = (ext: string): string => {
    switch (ext) {
        case "mp3":
            return "mpeg";
        case "svg":
            return "svg+xml"
        default:
            return ext;
    }
}

const toggleFullscreen = () => {
    const container = document.getElementById("vault-file-content");

    pdfFullScreen = !pdfFullScreen;

    if (!document.fullscreenElement) {
        // Enter fullscreen
        if (container.requestFullscreen) {
            container.requestFullscreen();
        } else if ((container as any).mozRequestFullScreen) {
            (container as any).mozRequestFullScreen(); // Firefox
        } else if ((container as any).webkitRequestFullscreen) {
            (container as any).webkitRequestFullscreen(); // Chrome, Safari, Opera
        } else if ((container as any).msRequestFullscreen) {
            (container as any).msRequestFullscreen(); // IE/Edge
        }
    } else {
        // Exit fullscreen
        if (document.exitFullscreen) {
            document.exitFullscreen();
        } else if ((document as any).mozCancelFullScreen) {
            (document as any).mozCancelFullScreen(); // Firefox
        } else if ((document as any).webkitExitFullscreen) {
            (document as any).webkitExitFullscreen(); // Chrome, Safari, Opera
        } else if ((document as any).msExitFullscreen) {
            (document as any).msExitFullscreen(); // IE/Edge
        }
    }

    queueRenderPage(pageNum);
}

/**
 * Renders a specific page of the currently loaded PDF
 * @param num - the page number to render
 */
const renderPDFPage = (num: number): void => {
    if (num < 1 || num > pdfDoc.numPages) {
        return;
    }

    pageNum = num;
    pageRendering = true;
    pdfDoc.getPage(num).then(function(page) {
        let scale = 2.25;
        if (pdfFullScreen) {
            scale += 1.25;
        }

        let viewport = page.getViewport({scale: scale});
        pdfCanvas.height = viewport.height;
        pdfCanvas.width = viewport.width;

        // Render PDF page into canvas context
        let renderContext = {
            canvasContext: pdfCtx,
            viewport: viewport
        };
        let renderTask = page.render(renderContext);

        // Wait for rendering to finish
        renderTask.promise.then(() => {
            pageRendering = false;
            if (pageNumPending !== null) {
                // New page rendering is pending
                renderPDFPage(pageNumPending);
                pageNumPending = null;
            }
        });
    });

    // Update page counters
    (document.getElementById(pageInputID) as HTMLInputElement).value = String(num);
    (document.getElementById(pageSpanID) as HTMLSpanElement).innerText = `/ ${pdfDoc.numPages}`;
}

/**
 * If another page rendering in progress, waits until the rendering is
 * finished. Otherwise, executes rendering immediately.
 * @param num - the new page number to render
 */
const queueRenderPage = (num: number) => {
    if (pageRendering) {
        pageNumPending = num;
    } else {
        renderPDFPage(num);
    }
}

/**
 * Displays previous page.
 */
const onPrevPage = () => {
    if (pageNum <= 1) {
        return;
    }
    pageNum--;
    queueRenderPage(pageNum);
}

/**
 * Displays next page.
 */
const onNextPage = () => {
    if (pageNum >= pdfDoc.numPages) {
        return;
    }
    pageNum++;
    queueRenderPage(pageNum);
}

/**
 * Uses pdf.js to render a PDF into a canvas from a decrypted Uint8Array
 * @param bytes - the decrypted pdf file bytes
 * @param callback - the vault view callback for inserting the element
 */
const renderPDF = (bytes: Uint8Array, callback: (string, MediaSource) => void) => {
    callback(pdfCanvasHTML, undefined);

    pageNum = 1;
    pdfjsLib.getDocument({
        data: bytes,
        isEvalSupported: false,
    }).promise.then(pdf => {
        pdfCanvas = document.getElementById(pdfCanvasID) as HTMLCanvasElement;
        pdfDoc = pdf;
        pdfCtx = pdfCanvas.getContext("2d");

        let fullscreen = document.getElementById(fullscreenID) as HTMLButtonElement;
        fullscreen.addEventListener("click", toggleFullscreen);

        document.getElementById(nextPageID).addEventListener("click", onNextPage);
        document.getElementById(prevPageID).addEventListener("click", onPrevPage);

        let numberInput = document.getElementById(pageInputID) as HTMLInputElement;
        numberInput.addEventListener("input", (event) => {
            queueRenderPage(Number((event.target as HTMLInputElement).value));
        });

        renderPDFPage(pageNum);
    });
}

/**
 * Renders a decrypted file into the vault view (if the file ext is supported)
 * @param filename - the filename (including extension)
 * @param bytes - the decrypted file bytes
 * @param callback - the vault view callback for inserting the element
 */
export const renderFileHTML = (filename: string, bytes: Uint8Array, callback: (string, MediaSource) => void) => {
    let ext = getExt(filename);
    if (ext === "pdf") {
        return renderPDF(bytes, callback);
    }

    let mime = getMimetypeFromExt(ext);
    let src = "REPLACE_SRC";

    let type;
    let tag;
    switch (ext) {
        case "jpeg":
        case "jpg":
        case "png":
        case "svg":
            type = {type: `image/${mime}`};
            tag = `<img src="${src}" alt="${filename}"/>`;
            break;
        case "mp3":
        case "wav":
            type = {type: `audio/${mime}`};
            tag = `<audio controls><source src="${src}" type="audio/${mime}">Unsupported format</audio>`;
            break;
        case "mp4":
        case "webm":
            type = {type: `video/${mime}`}
            tag = `<video controls><source src="${src}" type="video/${mime}">Unsupported format</video>`;
            break;
    }

    let blob = new Blob([bytes], type);
    let blobURL = URL.createObjectURL(blob);
    tag = tag.replace(src, blobURL);
    callback(tag, blobURL);
}

export const isNonTextFileType = (filename: string) => {
    let ext = getExt(filename);
    return nonTextFileTypes.includes(ext);
}