##Functional Requirements

The company aims to establish and maintain ownership of its infrastructure due to concerns over data privacy and the potential cost escalations associated with managed GenAI services. By investing in their own infrastructure, they aim to have greater control over data security and long-term financial sustainability.

To achieve this, they plan to invest in an AI PC with a budget of $10,000 - $15,000. This infrastructure must efficiently support 300 active students, all of whom are located within the city of Nagasaki. The solution should ensure reliability, minimal latency, and cost-effectiveness while maintaining performance requirements.

##Assumptions

The selected open-source Large Language Models (LLMs) will be sufficiently powerful to run efficiently within the allocated $10,000 - $15,000 hardware budget.

The existing network infrastructure, utilizing a single server connected to the internet from the office, will provide sufficient bandwidth and reliability to serve the needs of 300 students concurrently.

Software optimizations, such as quantization and model fine-tuning, will be leveraged to maximize computational efficiency.

Hardware procurement and setup will be feasible within the available budget, without exceeding operational constraints.

##Data Strategy

Given concerns over the use of copyrighted materials, all necessary learning materials will be legally procured and securely stored in an internal database.

Proper data governance policies will be established to ensure compliance with copyright regulations and ethical AI usage guidelines.

The data infrastructure will include mechanisms for efficient retrieval, version control, and access management to maintain integrity and scalability.

##Considerations

IBM Granite is the preferred LLM due to its fully open-source nature, transparent training data, and adherence to traceability standards. This selection ensures compliance with copyright regulations while maintaining visibility into model behavior and data origins.

The chosen AI PC must be optimized to handle both inference and potential fine-tuning workloads efficiently.

Future scalability considerations will include potential multi-GPU configurations or distributed computing strategies to accommodate increasing workloads.

Security measures, such as network segmentation and access control policies, will be implemented to protect sensitive student data and model outputs.

More details on IBM Granite: IBM Granite on Hugging Face
https://huggingface.co/ibm-granite
