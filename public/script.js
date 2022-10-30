
let forward = true;
let timer = 2;

function setActiveStep(step) {

    document.querySelectorAll('#example img').forEach((img) => img.removeAttribute('class'));
    document.querySelectorAll('#example img[data-step="' + step + '"]')[0].setAttribute('class', 'active');


    // // progress bar
    document.querySelectorAll('.progress .bar')[0].setAttribute('style', '--step:' + step);

    // // info
    document.querySelectorAll('#info .info').forEach((img) => img.classList.remove('active'));
    document.querySelectorAll('#info .info.step' + step)[0].classList.add('active');
}

function switchExample() {

    if (timer) {
        timer--;
        return;
    }


    const currentStep = parseInt(document.querySelectorAll('#info .info.active')[0].getAttribute('data-step'));
    const nextStep = forward ? (currentStep + 1) % 8 : Math.abs((currentStep - 1) % 8);
    if (nextStep === 0 || nextStep === 7) {
        forward = !forward;
        timer = 2;
    }
    setActiveStep(nextStep);
}

setInterval(switchExample, 1200);

function onIndicatorClick(e) {
    e.preventDefault();
    const step = e.currentTarget.getAttribute('data-step');
    document.querySelectorAll('aside')[0].classList.add('animation-disabled');
    timer = 1;
    setActiveStep(step);
    setTimeout(() => { document.querySelectorAll('aside')[0].classList.remove('animation-disabled'); }, 200);

}

document.querySelectorAll('span.indicator').forEach((span) => span.addEventListener('click', onIndicatorClick));