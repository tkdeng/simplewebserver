"use strict";

;(function(){
  if (WebAssembly) {
    // WebAssembly.instantiateStreaming is not currently available in Safari
    if (!WebAssembly.instantiateStreaming) {
      // polyfill
      WebAssembly.instantiateStreaming = async (resp, importObject) => {
        const source = await (await resp).arrayBuffer();
        return await WebAssembly.instantiate(source, importObject);
      };
    }

    // listen for scripts with type="wasm/go"
    setInterval(function(){
      document.querySelectorAll('script[type="wasm/go"]:not([gowasmloaded])').forEach(function(elm){
        elm.setAttribute('gowasmloaded', '');

        const src = elm.src;
        if(src.startsWith(window.location.origin+'/')){
          const go = new Go();
          WebAssembly.instantiateStreaming(fetch(src), go.importObject).then((result) => {
            go.run(result.instance);
          });
        }
      });
    }, 100);
  }else{
    console.error('WebAssembly is not supported in your browser');
  }
})();
