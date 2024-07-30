let counter = 1;
customElements.define(
    "exporter-config",
    class extends HTMLElement {
        constructor() {
            super();
        }

        connectedCallback() {
            const idx = counter++;

            const fieldSet = document.createElement("fieldset");
            fieldSet.classList.add("grid");

            const createInput = (title, name, def) => {
                const label = document.createElement("label");
                label.appendChild(document.createTextNode(title));
                const i = document.createElement("input");
                i.setAttribute("type", "text");
                i.classList.add("field-" + name);
                i.setAttribute("name", name + "_" + idx);
                const value = this.getAttribute(name) || def || "";
                i.setAttribute("value", value);
                label.appendChild(i);
                return label;
            };

            fieldSet.appendChild(createInput("Storage for exports", "storage"));
            fieldSet.appendChild(createInput("Item", "item"));
            fieldSet.appendChild(createInput("Slot", "slot"));
            fieldSet.appendChild(createInput("Amount", "amount", "64"));

            const removeBtn = document.createElement("button");
            removeBtn.innerHTML = "Del";
            removeBtn.addEventListener("click", () => {
                this.remove();
            });
            fieldSet.appendChild(removeBtn);
            this.appendChild(fieldSet);
        }
    },
);

customElements.define(
    "exporter-configs",
    class extends HTMLElement {
        constructor() {
            super();

            this.querySelector(".add-worker").addEventListener("click", (e) => {
                e.preventDefault();
                const worker = document.createElement("exporter-config");
                this.appendChild(worker);
            });
        }
    },
);
