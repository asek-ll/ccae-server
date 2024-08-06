(function () {
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

                fieldSet.appendChild(
                    createInput("Storage for exports", "storage"),
                );
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
        "importer-config",
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

                fieldSet.appendChild(
                    createInput("Storage for import", "storage"),
                );
                fieldSet.appendChild(createInput("Slot", "slot"));

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

                this.querySelector(".add-worker").addEventListener(
                    "click",
                    (e) => {
                        e.preventDefault();
                        const worker =
                            document.createElement("exporter-config");
                        this.appendChild(worker);
                    },
                );
            }
        },
    );

    customElements.define(
        "importer-configs",
        class extends HTMLElement {
            constructor() {
                super();

                this.querySelector(".add-worker").addEventListener(
                    "click",
                    (e) => {
                        e.preventDefault();
                        const worker =
                            document.createElement("importer-config");
                        this.appendChild(worker);
                    },
                );
            }
        },
    );

    /*
<details class="dropdown">
  <summary>Dropdown</summary>
  <ul>
    <li><a href="#">Solid</a></li>
    <li><a href="#">Liquid</a></li>
    <li><a href="#">Gas</a></li>
    <li><a href="#">Plasma</a></li>
  </ul>
</details>
     */

    const itemSelectorTemplate = `
	<dialog open id="item-select">
		<article>
			<h2>Items</h2>
			<form>
				<fieldset>
					<label>
						Name
						<input name="filter" />
					</label>
				</fieldset>
			</form>
			<div id="item-popup-items"></div>
			<footer>
				<button class="cancel-btn">Cancel</button>
			</footer>
		</article>
	</dialog>
    `;

    let dialogElement = null;

    customElements.define(
        "item-select-dialog",
        class extends HTMLElement {
            constructor() {
                super();
                this.addEventListener("show-dialog", (e) => {
                    const [resolve, reject] = e.detail;
                    this.resolve = resolve;
                    this.reject = reject;
                    this.style.display = "block";
                });
            }
            connectedCallback() {
                this.style.display = "none";
                const fragment = document
                    .createRange()
                    .createContextualFragment(itemSelectorTemplate);

                this.appendChild(fragment);

                this.querySelector(".cancel-btn").addEventListener(
                    "click",
                    () => {
                        if (this.reject != null) {
                            this.reject();
                        }
                        this.reject = null;
                        this.resolve = null;
                        this.style.display = "none";
                    },
                );
            }
        },
    );

    function selectItem() {
        if (dialogElement == null) {
            const dialog = document.createElement("item-select-dialog");
            document.body.appendChild(dialog);
            dialogElement = dialog;
        }
        return new Promise((resolve, reject) => {
            const event = new CustomEvent("show-dialog", {
                detail: [resolve, reject],
            });
            dialogElement.dispatchEvent(event);
        });
    }

    customElements.define(
        "item-selector",
        class extends HTMLElement {
            constructor() {
                super();

                // const details = document.createElement("details");
                // details.classList.add("dropdown");
                // const summary = document.createElement("summary");
                // const i = document.createElement("input");
                // summary.appendChild(i);

                // const ul = document.createElement("ul");
                // details.appendChild(summary);
                // details.appendChild(ul);

                // const li = document.createElement("li");
                // li.innerHTML = "TEST";
                // ul.appendChild(li);

                const btn = document.createElement("button");
                btn.innerHTML = "Sel";
                btn.addEventListener("click", (e) => {
                    e.preventDefault();
                    selectItem().then((i) => console.log(i));
                });

                this.appendChild(btn);
            }
        },
    );
})();
