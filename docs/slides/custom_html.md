```{raw} html
<style type="text/css">
    /* 1. Style header/footer <div> so they are positioned as desired. */
    #header-left {
        position: absolute;
        top: 0%;
        left: 0%;
    }
    #header-right {
        position: absolute;
        top: 0%;
        right: 0%;
    }
    #footer-left {
        position: absolute;
        bottom: 0%;
        left: 0%;
    }
</style>
<div id="hidden" style="display:none;">
    <div id="header">
        <div id="header-left"></div>
        <div id="header-right"></div>
        <div id="footer-left"></div>
    </div>
</div>
<script src="_static/jquery.js"></script>
<script type="text/javascript">
    var header = $('#header').html();
    if ( window.location.search.match( /print-pdf/gi ) ) {
        Reveal.addEventListener( 'ready', function( event ) {
            $('.slide-background').append(header);
        });
    }
    else {
        $('div.reveal').append(header);
   }
</script>
```
